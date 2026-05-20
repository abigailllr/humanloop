import math
import cv2
import mediapipe as mp
from pathlib import Path

from .vision import detect, hand_contacts
from .gemini import validate as gemini_validate
from .synthetic import check_metadata, check_motion_naturalness
from .kinematics import compute_joint_angles, to_motor_state, build_observation_vector
from ..robots.config import RobotConfig, ROBOTS, get_robot

_BLUR_THRESHOLD = 80.0
_CONFIDENCE_THRESHOLD = 0.5
_L_WRIST = 15
_R_WRIST = 16
_MAX_WRIST_SPEED_PER_SEC = 0.5


def _frame_quality_flags(gray_frame, pose_landmarks):
    blur_score = cv2.Laplacian(gray_frame, cv2.CV_64F).var()
    motion_blur = blur_score < _BLUR_THRESHOLD

    if pose_landmarks is None:
        low_confidence = True
    else:
        avg_vis = sum(lm["v"] for lm in pose_landmarks) / len(pose_landmarks)
        low_confidence = avg_vis < _CONFIDENCE_THRESHOLD

    flags = {}
    if motion_blur:
        flags["motion_blur"] = True
    if low_confidence:
        flags["low_confidence"] = True
    return flags


def _task_completion_score(frames, fps):
    if not frames or fps == 0:
        return 0.0
    total_disp = 0.0
    for i in range(1, len(frames)):
        prev = frames[i - 1].get("pose")
        curr = frames[i].get("pose")
        if not prev or not curr:
            continue
        for idx in (_L_WRIST, _R_WRIST):
            if idx < len(prev) and idx < len(curr):
                dx = curr[idx]["x"] - prev[idx]["x"]
                dy = curr[idx]["y"] - prev[idx]["y"]
                total_disp += math.sqrt(dx * dx + dy * dy)
    duration = len(frames) / fps
    normalized = total_disp / (duration + 1e-9) / _MAX_WRIST_SPEED_PER_SEC
    return min(1.0, round(normalized, 4))

HMDF_VERSION = "1.8"

mp_pose = mp.solutions.pose
mp_hands = mp.solutions.hands


def _velocity(prev, curr, dt):
    if not prev or not curr or dt == 0:
        return None
    return [
        {
            "vx": round((c["x"] - p["x"]) / dt, 6),
            "vy": round((c["y"] - p["y"]) / dt, 6),
            "vz": round((c["z"] - p["z"]) / dt, 6),
        }
        for p, c in zip(prev, curr)
    ]


def _contact_events(frames, fps):
    active = {}
    events = []
    n = len(frames)
    for i, frame in enumerate(frames):
        current = {(c["hand"], c["object_label"]) for c in frame.get("contacts", [])}
        for key in current - active.keys():
            active[key] = i
        for key in list(active.keys()):
            if key not in current:
                start = active.pop(key)
                events.append({"hand": key[0], "object_label": key[1], "t_start": round(start / fps, 4), "t_end": round(i / fps, 4)})
    for key, start in active.items():
        events.append({"hand": key[0], "object_label": key[1], "t_start": round(start / fps, 4), "t_end": round((n - 1) / fps, 4)})
    return events


def _stats(frames):
    touched = set()
    hand_frames = [0, 0]
    bimanual = False
    for frame in frames:
        hands = frame.get("hands") or []
        if len(hands) >= 2:
            bimanual = True
        for c in frame.get("contacts", []):
            touched.add(c["object_label"])
            if c["hand"] < 2:
                hand_frames[c["hand"]] += 1
    dominant = None
    if hand_frames[0] > hand_frames[1]:
        dominant = "right"
    elif hand_frames[1] > hand_frames[0]:
        dominant = "left"
    return {"bimanual": bimanual, "objects_touched": list(touched), "dominant_hand": dominant}


def extract(
    video_path: str,
    submission_id: str = "",
    challenge_id: str = "",
    challenge_title: str = "",
    user_id: str = "",
    lat: float = 0.0,
    lng: float = 0.0,
    captured_at: str = "",
    robot: str = "generic_bimanual",
) -> dict:
    robot_cfg: RobotConfig = get_robot(robot)

    cap = cv2.VideoCapture(video_path)
    fps = cap.get(cv2.CAP_PROP_FPS)
    dt = 1.0 / fps if fps > 0 else 0
    frames = []
    frame_index = 0
    prev_pose = None
    prev_hands = None
    prev_joint_angles = None
    prev_motor = None

    with mp_pose.Pose() as pose_model, mp_hands.Hands() as hands_model:
        while cap.isOpened():
            ok, frame = cap.read()
            if not ok:
                break
            gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            rgb = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
            pose_result = pose_model.process(rgb)
            hands_result = hands_model.process(rgb)

            pose_landmarks = None
            if pose_result.pose_landmarks:
                pose_landmarks = [
                    {"x": lm.x, "y": lm.y, "z": lm.z, "v": lm.visibility}
                    for lm in pose_result.pose_landmarks.landmark
                ]

            hand_landmarks = []
            hand_labels = []
            if hands_result.multi_hand_landmarks:
                for i, hand in enumerate(hands_result.multi_hand_landmarks):
                    hand_landmarks.append([{"x": lm.x, "y": lm.y, "z": lm.z} for lm in hand.landmark])
                    if hands_result.multi_handedness:
                        hand_labels.append(hands_result.multi_handedness[i].classification[0].label)

            objects = detect(rgb)
            contacts = hand_contacts(hand_landmarks, objects)

            pose_vel = _velocity(prev_pose, pose_landmarks, dt) if pose_landmarks else None
            hand_vel = [
                _velocity(prev_hands[i] if prev_hands and i < len(prev_hands) else None, hand_landmarks[i], dt)
                for i in range(len(hand_landmarks))
            ] or None

            joint_angles = compute_joint_angles(pose_landmarks) if pose_landmarks else {}
            motor = to_motor_state(
                joint_angles,
                prev_joint_angles if frame_index > 0 else None,
                dt,
                prev_motor.get("dq") if prev_motor else None,
                robot_cfg,
                hand_landmarks if robot_cfg.has_gripper else None,
            )
            t_sec = round(frame_index / fps, 4) if fps > 0 else 0
            obs = build_observation_vector(motor, prev_motor, t_sec, robot_cfg)

            quality_flags = _frame_quality_flags(gray, pose_landmarks)

            entry = {
                "t": t_sec,
                "pose": pose_landmarks,
                "joint_angles": joint_angles,
                "motor_state": motor,
                "obs": obs,
                "hands": hand_landmarks,
                "objects": objects,
                "contacts": contacts,
            }
            if quality_flags:
                entry["quality_flags"] = quality_flags
            if pose_vel:
                entry["pose_vel"] = pose_vel
            if hand_vel:
                entry["hand_vel"] = hand_vel
            if hand_labels:
                entry["hand_labels"] = hand_labels

            frames.append(entry)
            prev_pose = pose_landmarks
            prev_hands = hand_landmarks
            prev_joint_angles = joint_angles
            prev_motor = motor
            frame_index += 1

    cap.release()

    task_completion = _task_completion_score(frames, fps)
    validation = gemini_validate(video_path, challenge_title)
    contact_events = _contact_events(frames, fps) if fps > 0 else []
    stats = _stats(frames)

    meta_check = check_metadata(video_path)
    motion_check = check_motion_naturalness(frames)

    gemini_synthetic = validation.get("synthetic", False)
    gemini_synthetic_confidence = validation.get("synthetic_confidence", 0.0)
    gemini_synthetic_signals = validation.get("synthetic_signals", [])

    all_signals = list(gemini_synthetic_signals) + meta_check.get("signals", []) + motion_check.get("signals", [])
    is_synthetic = (
        gemini_synthetic
        or gemini_synthetic_confidence > 0.7
        or meta_check.get("suspicious") and len(meta_check.get("signals", [])) >= 2
        or motion_check.get("suspicious") and len(motion_check.get("signals", [])) >= 2
    )

    return {
        "hmdf_version": HMDF_VERSION,
        "source": "humanloop",
        "submission_id": submission_id,
        "challenge_id": challenge_id,
        "challenge_title": challenge_title,
        "user_id": user_id,
        "robot": robot,
        "fps": fps,
        "frame_count": len(frames),
        "frames": frames,
        "contact_events": contact_events,
        "stats": stats,
        "task_completion_score": task_completion,
        "validation": validation,
        "synthetic_detection": {
            "synthetic": is_synthetic,
            "signals": all_signals,
            "gemini_confidence": gemini_synthetic_confidence,
            "metadata": meta_check,
            "motion": motion_check,
        },
        "metadata": {
            "task_type": "manipulation",
            "coordinate_space": "normalized",
            "pose_landmarks": 33,
            "hand_landmarks": 21,
            "vision_model": "yolo11x",
            "joint_angles_standard": "anatomical",
            "joint_angles_unit": "degrees",
            "robot_type": robot_cfg.robot_type,
            "dof": robot_cfg.dof,
            "has_gripper": robot_cfg.has_gripper,
            "obs_dim": robot_cfg.obs_dim,
            "action_dim": robot_cfg.action_dim,
            "joint_names": [j.robot_joint for j in robot_cfg.joints],
        },
        **({"location": {"lat": lat, "lng": lng}} if lat != 0.0 or lng != 0.0 else {}),
        **({"captured_at": captured_at} if captured_at else {}),
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
