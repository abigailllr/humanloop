import gzip
import json
import sys
import os
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq
from pathlib import Path


def flatten_frame(frame: dict, submission_id: str, challenge_id: str, user_id: str) -> list[dict]:
    rows = []
    t = frame.get("t", 0)
    pose = frame.get("pose") or []
    hands = frame.get("hands") or []

    for i, lm in enumerate(pose):
        rows.append({
            "submission_id": submission_id,
            "challenge_id": challenge_id,
            "user_id": user_id,
            "t": t,
            "source": "pose",
            "landmark_index": i,
            "x": lm["x"],
            "y": lm["y"],
            "z": lm["z"],
            "v": lm.get("v", 1.0),
        })

    for hand_index, hand in enumerate(hands):
        for i, lm in enumerate(hand):
            rows.append({
                "submission_id": submission_id,
                "challenge_id": challenge_id,
                "user_id": user_id,
                "t": t,
                "source": f"hand_{hand_index}",
                "landmark_index": i,
                "x": lm["x"],
                "y": lm["y"],
                "z": lm["z"],
                "v": 1.0,
            })

    return rows


def convert(input_dir: str, output_path: str):
    rows = []
    for path in Path(input_dir).glob("*.hmdf.json.gz"):
        with gzip.open(path, "rt") as f:
            record = json.load(f)

        sub_id = record.get("submission_id", "")
        ch_id = record.get("challenge_id", "")
        u_id = record.get("user_id", "")

        for frame in record.get("frames", []):
            rows.extend(flatten_frame(frame, sub_id, ch_id, u_id))

    if not rows:
        print("no data found")
        return

    df = pd.DataFrame(rows)
    table = pa.Table.from_pandas(df)
    pq.write_table(table, output_path, compression="snappy")
    print(f"exported {len(rows)} rows to {output_path}")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("usage: export.py <extracted_dir> <output.parquet>")
        sys.exit(1)
    convert(sys.argv[1], sys.argv[2])
