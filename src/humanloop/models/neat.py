import gzip
import json
import os
import pickle
import tempfile
from pathlib import Path

import neat


_NEAT_CONFIG_TEMPLATE = """
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
num_inputs              = {obs_dim}
num_outputs             = {action_dim}
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


def _detect_dims(record: dict) -> tuple[int, int]:
    for frame in record.get("frames", []):
        obs = frame.get("obs")
        q = frame.get("motor_state", {}).get("q")
        if obs and q:
            return len(obs), len(q)
    meta = record.get("metadata", {})
    obs_dim = meta.get("obs_dim", 32)
    action_dim = meta.get("action_dim", 10)
    return obs_dim, action_dim


def _load_dataset(data_dir: str, challenge_id: str = "") -> tuple[list[tuple[list, list]], int, int]:
    pairs = []
    obs_dim = action_dim = 0

    for path in Path(data_dir).glob("*.gz"):
        try:
            record = _load_hmdf(str(path))
        except Exception:
            continue
        if challenge_id and record.get("challenge_id") != challenge_id:
            continue
        if record.get("synthetic_detection", {}).get("synthetic"):
            continue
        if record.get("validation", {}).get("confidence", 0.0) < 0.5:
            continue

        if obs_dim == 0:
            obs_dim, action_dim = _detect_dims(record)

        frames = record.get("frames", [])
        for i, frame in enumerate(frames):
            obs = frame.get("obs")
            if not obs or len(obs) != obs_dim:
                continue
            next_ms = frames[i + 1].get("motor_state", {}) if i + 1 < len(frames) else frame.get("motor_state", {})
            action = next_ms.get("q", [])
            if len(action) != action_dim:
                continue
            pairs.append((obs, action))

    return pairs, obs_dim, action_dim


def _mse(predicted: list, target: list) -> float:
    return sum((p - t) ** 2 for p, t in zip(predicted, target)) / len(target)


def _evaluate(genomes, config, dataset):
    sample = dataset[:2000] if len(dataset) > 2000 else dataset
    for gid, genome in genomes:
        net = neat.nn.FeedForwardNetwork.create(genome, config)
        total = sum(_mse(net.activate(obs), action) for obs, action in sample)
        genome.fitness = -(total / max(len(sample), 1))


def _make_config(obs_dim: int, action_dim: int) -> neat.Config:
    cfg_str = _NEAT_CONFIG_TEMPLATE.format(obs_dim=obs_dim, action_dim=action_dim)
    with tempfile.NamedTemporaryFile(mode="w", suffix=".cfg", delete=False) as tmp:
        tmp.write(cfg_str)
        cfg_path = tmp.name
    try:
        return neat.Config(
            neat.DefaultGenome,
            neat.DefaultReproduction,
            neat.DefaultSpeciesSet,
            neat.DefaultStagnation,
            cfg_path,
        )
    finally:
        os.unlink(cfg_path)


def train(data_dir: str, out_path: str, generations: int = 50, challenge_id: str = "") -> None:
    dataset, obs_dim, action_dim = _load_dataset(data_dir, challenge_id)
    if not dataset:
        raise ValueError("no valid training data found")

    config = _make_config(obs_dim, action_dim)
    pop = neat.Population(config)
    pop.add_reporter(neat.StdOutReporter(True))
    pop.add_reporter(neat.StatisticsReporter())

    winner = pop.run(lambda genomes, cfg: _evaluate(genomes, cfg, dataset), generations)

    with open(out_path, "wb") as f:
        pickle.dump((winner, obs_dim, action_dim), f)


def run(genome_path: str, obs: list[float]) -> list[float]:
    with open(genome_path, "rb") as f:
        payload = pickle.load(f)

    if isinstance(payload, tuple):
        genome, obs_dim, action_dim = payload
    else:
        genome, obs_dim, action_dim = payload, 32, 10

    config = _make_config(obs_dim, action_dim)
    net = neat.nn.FeedForwardNetwork.create(genome, config)
    return list(net.activate(obs))
