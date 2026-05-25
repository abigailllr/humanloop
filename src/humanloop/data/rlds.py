import gzip
import json
from pathlib import Path


def _load_hmdf(path: str) -> dict:
    if path.endswith(".gz"):
        with gzip.open(path, "rt", encoding="utf-8") as f:
            return json.load(f)
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def export(hmdf_paths: list[str], out_dir: str) -> None:
    out = Path(out_dir)
    out.mkdir(parents=True, exist_ok=True)

    with open(out / "episodes.jsonl", "w") as ef:
        for hmdf_path in hmdf_paths:
            record = _load_hmdf(hmdf_path)
            frames = record.get("frames", [])
            if not frames:
                continue

            n = len(frames)
            fps = record.get("fps") or 30.0
            steps = []

            for i, frame in enumerate(frames):
                ms = frame.get("motor_state", {})
                obs_vec = frame.get("obs", [])
                q = ms.get("q", [])
                next_q = frames[i + 1].get("motor_state", {}).get("q", q) if i + 1 < n else q

                steps.append({
                    "observation": {"state": obs_vec},
                    "action": next_q,
                    "reward": 1.0 if i == n - 1 else 0.0,
                    "discount": 1.0,
                    "is_first": i == 0,
                    "is_last": i == n - 1,
                    "is_terminal": i == n - 1,
                })

            ef.write(json.dumps({
                "episode_metadata": {
                    "submission_id": record.get("submission_id", ""),
                    "challenge_id": record.get("challenge_id", ""),
                    "challenge_title": record.get("challenge_title", ""),
                    "robot": record.get("robot", ""),
                    "fps": fps,
                    "task_completion_score": record.get("task_completion_score", 0.0),
                },
                "steps": steps,
            }) + "\n")
