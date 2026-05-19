import cv2
import mediapipe as mp
import json
import sys
from pathlib import Path
from vision import detect, hand_contacts

HMDF_VERSION = "1.2"

mp_pose = mp.solutions.pose
mp_hands = mp.solutions.hands


def extract(video_path: str, submission_id: str = "", challenge_id: str = "", challenge_title: str = "", user_id: str = "") -> dict:
    cap = cv2.VideoCapture(video_path)
    fps = cap.get(cv2.CAP_PROP_FPS)
    frames = []
    frame_index = 0

    with mp_pose.Pose() as pose, mp_hands.Hands() as hands:
        while cap.isOpened():
            ok, frame = cap.read()
            if not ok:
                break
            rgb = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
            pose_result = pose.process(rgb)
            hands_result = hands.process(rgb)

            pose_landmarks = None
            if pose_result.pose_landmarks:
                pose_landmarks = [
                    {"x": lm.x, "y": lm.y, "z": lm.z, "v": lm.visibility}
                    for lm in pose_result.pose_landmarks.landmark
                ]

            hand_landmarks = []
            if hands_result.multi_hand_landmarks:
                for hand in hands_result.multi_hand_landmarks:
                    hand_landmarks.append(
                        [{"x": lm.x, "y": lm.y, "z": lm.z} for lm in hand.landmark]
                    )

            objects = detect(rgb)
            contacts = hand_contacts(hand_landmarks, objects)

            frames.append({
                "t": round(frame_index / fps, 4) if fps > 0 else 0,
                "pose": pose_landmarks,
                "hands": hand_landmarks,
                "objects": objects,
                "contacts": contacts,
            })
            frame_index += 1

    cap.release()

    return {
        "hmdf_version": HMDF_VERSION,
        "source": "humanloop",
        "submission_id": submission_id,
        "challenge_id": challenge_id,
        "challenge_title": challenge_title,
        "user_id": user_id,
        "fps": fps,
        "frame_count": len(frames),
        "frames": frames,
        "metadata": {
            "task_type": "manipulation",
            "coordinate_space": "normalized",
            "pose_landmarks": 33,
            "hand_landmarks": 21,
            "vision_model": "yolo11x",
        },
    }


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
        print("usage: main.py <validate|extract> <video_path> [submission_id] [challenge_id] [challenge_title] [user_id]")
        sys.exit(1)

    command = sys.argv[1]
    video = sys.argv[2]

    if command == "validate":
        print(json.dumps(validate(video)))
    elif command == "extract":
        sub_id = sys.argv[3] if len(sys.argv) > 3 else ""
        ch_id = sys.argv[4] if len(sys.argv) > 4 else ""
        ch_title = sys.argv[5] if len(sys.argv) > 5 else ""
        u_id = sys.argv[6] if len(sys.argv) > 6 else ""
        print(json.dumps(extract(video, sub_id, ch_id, ch_title, u_id)))
