import gzip
import json
import math
import os
import pickle
import sys
import tempfile
from pathlib import Path

import neat


OBS_DIM = 32
ACTION_DIM = 10

_NEAT_CONFIG = """
[NEAT]
fitness_criterion     = max
fitness_threshold     = -0.001
pop_size              = 150
reset_on_extinction   = False

[DefaultGenome]
activation_default      = tanh
activation_mutate_rate  = 0.0
activation_options      = tanh
aggregation_default     = sum
aggregation_mutate_rate = 0.0
aggregation_options     = sum
bias_init_mean          = 0.0
bias_init_stdev         = 1.0
bias_max_value          = 30.0
bias_min_value          = -30.0
bias_mutate_power       = 0.5
bias_mutate_rate        = 0.7
bias_replace_rate       = 0.1
compatibility_disjoint_coefficient = 1.0
compatibility_weight_coefficient   = 0.5
conn_add_prob           = 0.5
conn_delete_prob        = 0.5
enabled_default         = True
enabled_mutate_rate     = 0.01
feed_forward            = True
initial_connection      = full_direct
node_add_prob           = 0.2
node_delete_prob        = 0.2
num_hidden              = 0
num_inputs              = 32
num_outputs             = 10
response_init_mean      = 1.0
response_init_stdev     = 0.0
response_max_value      = 30.0
response_min_value      = -30.0
response_mutate_power   = 0.0
response_mutate_rate    = 0.0
response_replace_rate   = 0.0
weight_init_mean        = 0.0
weight_init_stdev       = 1.0
weight_max_value        = 30
weight_min_value        = -30
weight_mutate_power     = 0.5
weight_mutate_rate      = 0.8
weight_replace_rate     = 0.1

[DefaultSpeciesSet]
compatibility_threshold = 3.0

[DefaultStagnation]
species_fitness_func = max
max_stagnation       = 20
species_elitism      = 2

[DefaultReproduction]
elitism            = 2
survival_threshold = 0.2
"""


def _load_hmdf(path: str) -> dict:
    if path.endswith(".gz"):
        with gzip.open(path, "rt", encoding="utf-8") as f:
            return json.load(f)
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def _load_dataset(data_dir: str, challenge_id: str = "") -> list[tuple[list, list]]:
    pairs = []
    for path in Path(data_dir).glob("*.gz"):
        try:
            record = _load_hmdf(str(path))
        except Exception:
            continue
        if challenge_id and record.get("challenge_id") != challenge_id:
            continue
        if record.get("synthetic_detection", {}).get("synthetic"):
            continue
        conf = record.get("validation", {}).get("confidence", 0.0)
        if conf < 0.5:
            continue

        frames = record.get("frames", [])
        for i, frame in enumerate(frames):
            obs = frame.get("obs")
            if not obs or len(obs) != OBS_DIM:
                continue
            if i + 1 < len(frames):
                next_ms = frames[i + 1].get("motor_state", {})
            else:
                next_ms = frame.get("motor_state", {})
            action = next_ms.get("q", [])
            if len(action) != ACTION_DIM:
                continue
            pairs.append((obs, action))

    return pairs


def _mse(predicted: list, target: list) -> float:
    return sum((p - t) ** 2 for p, t in zip(predicted, target)) / len(target)


def _evaluate(genomes, config, dataset):
    nets = [(gid, neat.nn.FeedForwardNetwork.create(genome, config), genome) for gid, genome in genomes]
    sample = dataset[:2000] if len(dataset) > 2000 else dataset

    for gid, net, genome in nets:
        total_mse = 0.0
        for obs, action in sample:
            output = net.activate(obs)
            total_mse += _mse(output, action)
        genome.fitness = -(total_mse / max(len(sample), 1))


def train(data_dir: str, out_path: str, generations: int = 50, challenge_id: str = "") -> None:
    dataset = _load_dataset(data_dir, challenge_id)
    if not dataset:
        print("no valid training data found")
        sys.exit(1)

    print(f"loaded {len(dataset)} (obs, action) pairs from {data_dir}")

    with tempfile.NamedTemporaryFile(mode="w", suffix=".cfg", delete=False) as tmp:
        tmp.write(_NEAT_CONFIG)
        cfg_path = tmp.name

    try:
        config = neat.Config(
            neat.DefaultGenome,
            neat.DefaultReproduction,
            neat.DefaultSpeciesSet,
            neat.DefaultStagnation,
            cfg_path,
        )
    finally:
        os.unlink(cfg_path)

    pop = neat.Population(config)
    pop.add_reporter(neat.StdOutReporter(True))
    pop.add_reporter(neat.StatisticsReporter())

    winner = pop.run(lambda genomes, cfg: _evaluate(genomes, cfg, dataset), generations)

    with open(out_path, "wb") as f:
        pickle.dump(winner, f)

    print(f"best genome fitness: {winner.fitness:.6f}")
    print(f"saved to {out_path}")


def run(genome_path: str, obs: list[float]) -> list[float]:
    with tempfile.NamedTemporaryFile(mode="w", suffix=".cfg", delete=False) as tmp:
        tmp.write(_NEAT_CONFIG)
        cfg_path = tmp.name

    try:
        config = neat.Config(
            neat.DefaultGenome,
            neat.DefaultReproduction,
            neat.DefaultSpeciesSet,
            neat.DefaultStagnation,
            cfg_path,
        )
    finally:
        os.unlink(cfg_path)

    with open(genome_path, "rb") as f:
        genome = pickle.load(f)

    net = neat.nn.FeedForwardNetwork.create(genome, config)
    return list(net.activate(obs))


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("usage: neat_train.py train <data_dir> <out_genome.pkl> [generations] [challenge_id]")
        print("       neat_train.py run   <genome.pkl> <obs_json>")
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "train":
        d_dir = sys.argv[2]
        o_path = sys.argv[3]
        gens = int(sys.argv[4]) if len(sys.argv) > 4 else 50
        ch_id = sys.argv[5] if len(sys.argv) > 5 else ""
        train(d_dir, o_path, gens, ch_id)

    elif cmd == "run":
        g_path = sys.argv[2]
        obs_input = json.loads(sys.argv[3])
        result = run(g_path, obs_input)
        print(json.dumps(result))
