import json
import sys

from humanloop.models.neat import train, run

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("usage: train.py train <data_dir> <out_genome.pkl> [generations] [challenge_id]")
        print("       train.py run   <genome.pkl> <obs_json>")
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
