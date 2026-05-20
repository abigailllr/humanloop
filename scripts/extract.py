import json
import sys

from humanloop.extraction.pipeline import extract, validate

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("usage: extract.py <validate|extract> <video_path> [submission_id] [challenge_id] [challenge_title] [user_id] [lat] [lng] [captured_at] [robot]")
        sys.exit(1)

    command = sys.argv[1]
    video = sys.argv[2]

    if command == "validate":
        print(json.dumps(validate(video)))
    elif command == "extract":
        sub_id     = sys.argv[3]  if len(sys.argv) > 3  else ""
        ch_id      = sys.argv[4]  if len(sys.argv) > 4  else ""
        ch_title   = sys.argv[5]  if len(sys.argv) > 5  else ""
        u_id       = sys.argv[6]  if len(sys.argv) > 6  else ""
        lat        = float(sys.argv[7]) if len(sys.argv) > 7 else 0.0
        lng        = float(sys.argv[8]) if len(sys.argv) > 8 else 0.0
        captured_at = sys.argv[9] if len(sys.argv) > 9  else ""
        robot      = sys.argv[10] if len(sys.argv) > 10 else "generic_bimanual"
        print(json.dumps(extract(video, sub_id, ch_id, ch_title, u_id, lat, lng, captured_at, robot)))
