import cv2
import mediapipe as mp
import json
import sys
from pathlib import Path

mp_pose = mp.solutions.pose
mp_hands = mp.solutions.hands


def extract(video_path: str) -> dict:
    cap = cv2.VideoCapture(video_path)
    fps = cap.get(cv2.CAP_PROP_FPS)
    frames = []

    with mp_pose.Pose() as pose, mp_hands.Hands() as hands:
        while cap.isOpened():
            ok, frame = cap.read()
            if not ok:
                break
            rgb = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
            pose_result = pose.process(rgb)
            hands_result = hands.process(rgb)

            frame_data = {"pose": None, "hands": []}

            if pose_result.pose_landmarks:
                frame_data["pose"] = [
                    {"x": lm.x, "y": lm.y, "z": lm.z, "v": lm.visibility}
                    for lm in pose_result.pose_landmarks.landmark
                ]

            if hands_result.multi_hand_landmarks:
                for hand in hands_result.multi_hand_landmarks:
                    frame_data["hands"].append(
                        [{"x": lm.x, "y": lm.y, "z": lm.z} for lm in hand.landmark]
                    )

            frames.append(frame_data)

    cap.release()
    return {"fps": fps, "frame_count": len(frames), "frames": frames}


def validate(video_path: str) -> dict:
    path = Path(video_path)
    issues = []

    if not path.exists():
        issues.append("file_not_found")

    cap = cv2.VideoCapture(video_path)
    frame_count = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    fps = cap.get(cv2.CAP_PROP_FPS)
    duration = frame_count / fps if fps > 0 else 0
    cap.release()

    if duration < 2:
        issues.append("too_short")
    if duration > 60:
        issues.append("too_long")

    return {"valid": len(issues) == 0, "issues": issues, "duration": duration}


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("usage: main.py <validate|extract> <video_path>")
        sys.exit(1)

    command = sys.argv[1]
    video = sys.argv[2]

    if command == "validate":
        print(json.dumps(validate(video)))
    elif command == "extract":
        print(json.dumps(extract(video)))
