import gzip
import json
from pathlib import Path


def _load_hmdf(path: str) -> dict:
    if path.endswith(".gz"):
        with gzip.open(path, "rt", encoding="utf-8") as f:
            return json.load(f)
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def _obs_dim(record: dict) -> int:
    for frame in record.get("frames", []):
        obs = frame.get("obs")
        if obs:
            return len(obs)
    return 0


def _action_dim(record: dict) -> int:
    for frame in record.get("frames", []):
        ms = frame.get("motor_state", {})
        q = ms.get("q", [])
        if q:
            return len(q)
    return 0


def export(hmdf_paths: list[str], out_dir: str, task_description: str = "") -> None:
    out = Path(out_dir)
    data_dir = out / "data" / "chunk-000"
    meta_dir = out / "meta"
    data_dir.mkdir(parents=True, exist_ok=True)
    meta_dir.mkdir(parents=True, exist_ok=True)

    rows = []
    episode_index = 0
    total_frames = 0
    obs_dim = 0
    action_dim = 0

    for hmdf_path in hmdf_paths:
        record = _load_hmdf(hmdf_path)
        frames = record.get("frames", [])
        if not frames:
            continue

        if obs_dim == 0:
            obs_dim = _obs_dim(record)
        if action_dim == 0:
            action_dim = _action_dim(record)

        fps = record.get("fps", 30.0)

        for i, frame in enumerate(frames):
            ms = frame.get("motor_state", {})
            obs = frame.get("obs", [0.0] * obs_dim)
            q   = ms.get("q",   [0.0] * action_dim)
            dq  = ms.get("dq",  [0.0] * action_dim)
            ddq = ms.get("ddq", [0.0] * action_dim)

            next_q = frames[i + 1].get("motor_state", {}).get("q", q) if i + 1 < len(frames) else q

            rows.append({
                "episode_index":       episode_index,
                "frame_index":         total_frames,
                "timestamp":           frame.get("t", round(i / fps, 4)),
                "observation.state":   obs,
                "action":              next_q,
                "motor_state.q":       q,
                "motor_state.dq":      dq,
                "motor_state.ddq":     ddq,
                "task_index":          0,
            })
            total_frames += 1

        episode_index += 1

    try:
        import pyarrow as pa
        import pyarrow.parquet as pq

        schema = pa.schema([
            pa.field("episode_index",     pa.int64()),
            pa.field("frame_index",       pa.int64()),
            pa.field("timestamp",         pa.float32()),
            pa.field("observation.state", pa.list_(pa.float32())),
            pa.field("action",            pa.list_(pa.float32())),
            pa.field("motor_state.q",     pa.list_(pa.float32())),
            pa.field("motor_state.dq",    pa.list_(pa.float32())),
            pa.field("motor_state.ddq",   pa.list_(pa.float32())),
            pa.field("task_index",        pa.int64()),
        ])

        table = pa.table({
            "episode_index":     [r["episode_index"]     for r in rows],
            "frame_index":       [r["frame_index"]       for r in rows],
            "timestamp":         [r["timestamp"]         for r in rows],
            "observation.state": [r["observation.state"] for r in rows],
            "action":            [r["action"]            for r in rows],
            "motor_state.q":     [r["motor_state.q"]     for r in rows],
            "motor_state.dq":    [r["motor_state.dq"]    for r in rows],
            "motor_state.ddq":   [r["motor_state.ddq"]   for r in rows],
            "task_index":        [r["task_index"]        for r in rows],
        }, schema=schema)

        pq.write_table(table, data_dir / "train-00000-of-00001.parquet")
    except ImportError:
        with open(data_dir / "train.jsonl", "w") as f:
            for row in rows:
                f.write(json.dumps(row) + "\n")

    info = {
        "codebase_version": "v2.0",
        "robot_type":       "generic_arm",
        "total_episodes":   episode_index,
        "total_frames":     total_frames,
        "fps":              fps,
        "splits":           {"train": f"0:{episode_index}"},
        "features": {
            "observation.state": {"dtype": "float32", "shape": [obs_dim],    "names": None},
            "action":            {"dtype": "float32", "shape": [action_dim], "names": None},
            "motor_state.q":     {"dtype": "float32", "shape": [action_dim], "names": None},
            "motor_state.dq":    {"dtype": "float32", "shape": [action_dim], "names": None},
            "motor_state.ddq":   {"dtype": "float32", "shape": [action_dim], "names": None},
            "timestamp":         {"dtype": "float32", "shape": [1],          "names": None},
            "episode_index":     {"dtype": "int64",   "shape": [1],          "names": None},
            "frame_index":       {"dtype": "int64",   "shape": [1],          "names": None},
            "task_index":        {"dtype": "int64",   "shape": [1],          "names": None},
        },
        "tasks": [{"task_index": 0, "task": task_description}],
    }

    with open(meta_dir / "info.json", "w") as f:
        json.dump(info, f, indent=2)

    episodes = [
        {"episode_index": i, "tasks": [task_description], "length": sum(1 for r in rows if r["episode_index"] == i)}
        for i in range(episode_index)
    ]
    with open(meta_dir / "episodes.jsonl", "w") as f:
        for ep in episodes:
            f.write(json.dumps(ep) + "\n")

    tasks = [{"task_index": 0, "task": task_description}]
    with open(meta_dir / "tasks.jsonl", "w") as f:
        for t in tasks:
            f.write(json.dumps(t) + "\n")
