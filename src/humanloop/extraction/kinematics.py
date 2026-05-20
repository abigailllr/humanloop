import math

from ..robots.config import RobotConfig, ROBOTS


def _vec(a, b):
    return (b["x"] - a["x"], b["y"] - a["y"], b["z"] - a["z"])


def _dot(u, v):
    return sum(x * y for x, y in zip(u, v))


def _norm(v):
    return math.sqrt(sum(x * x for x in v))


def _angle_deg(u, v):
    d = _dot(u, v)
    n = _norm(u) * _norm(v)
    if n < 1e-9:
        return 0.0
    return round(math.degrees(math.acos(max(-1.0, min(1.0, d / n)))), 2)


def _cross(u, v):
    return (
        u[1] * v[2] - u[2] * v[1],
        u[2] * v[0] - u[0] * v[2],
        u[0] * v[1] - u[1] * v[0],
    )


def _signed_angle_deg(u, v, normal):
    angle = _angle_deg(u, v)
    cross = _cross(u, v)
    sign = 1 if _dot(cross, normal) >= 0 else -1
    return round(sign * angle, 2)


_MP = {
    "nose": 0,
    "l_shoulder": 11, "r_shoulder": 12,
    "l_elbow": 13,    "r_elbow": 14,
    "l_wrist": 15,    "r_wrist": 16,
    "l_hip": 23,      "r_hip": 24,
    "l_index": 19,    "r_index": 20,
    "l_pinky": 17,    "r_pinky": 18,
}


def compute_joint_angles(pose: list) -> dict:
    if not pose or len(pose) < 25:
        return {}

    def lm(name):
        return pose[_MP[name]]

    angles = {}

    for side in ("l", "r"):
        opp = "r" if side == "l" else "l"
        shoulder = lm(f"{side}_shoulder")
        elbow    = lm(f"{side}_elbow")
        wrist    = lm(f"{side}_wrist")
        hip      = lm(f"{side}_hip")
        opp_sh   = lm(f"{opp}_shoulder")

        trunk_axis = _vec(hip, shoulder)
        lateral    = _vec(shoulder, opp_sh) if side == "l" else _vec(opp_sh, shoulder)
        upper_arm  = _vec(shoulder, elbow)
        forearm    = _vec(elbow, wrist)

        shoulder_flex   = _signed_angle_deg(trunk_axis, upper_arm, lateral)
        shoulder_abduct = _angle_deg(lateral, upper_arm)
        elbow_flex      = 180.0 - _angle_deg(upper_arm, forearm)

        index    = lm(f"{side}_index")
        pinky    = lm(f"{side}_pinky")
        hand     = _vec(wrist, index)
        hand_lat = _vec(pinky, index)
        wrist_flex    = _signed_angle_deg(forearm, hand, hand_lat)
        wrist_deviate = _signed_angle_deg(forearm, hand, _cross(forearm, hand_lat))

        label = "left" if side == "l" else "right"
        angles[label] = {
            "shoulder_flexion":   shoulder_flex,
            "shoulder_abduction": shoulder_abduct,
            "elbow_flexion":      elbow_flex,
            "wrist_flexion":      wrist_flex,
            "wrist_deviation":    wrist_deviate,
        }

    return angles


def gripper_aperture(hand_landmarks: list[dict]) -> float:
    if not hand_landmarks or len(hand_landmarks) < 21:
        return 1.0

    def dist(a, b):
        return math.sqrt(
            (a["x"] - b["x"]) ** 2 +
            (a["y"] - b["y"]) ** 2 +
            (a["z"] - b["z"]) ** 2
        )

    hand_size = dist(hand_landmarks[0], hand_landmarks[9]) or 1e-6
    return round(min(1.0, dist(hand_landmarks[4], hand_landmarks[8]) / hand_size), 4)


def _clamp(val: float, lo: float, hi: float) -> float:
    return max(lo, min(hi, val))


def to_motor_state(
    joint_angles: dict,
    prev_angles: dict | None,
    dt: float,
    prev_dq: list | None = None,
    robot: RobotConfig | None = None,
    hand_landmarks: list[list[dict]] | None = None,
) -> dict:
    if robot is None:
        robot = ROBOTS["generic_bimanual"]

    prev_dq_cache = prev_dq if prev_dq and len(prev_dq) == robot.dof else [0.0] * robot.dof

    q, dq, ddq = [], [], []

    for i, jm in enumerate(robot.joints):
        if jm.human_side == "const":
            q_rad = jm.const_val
            dq_val = 0.0
        elif jm.human_side == "gripper":
            landmarks = None
            if hand_landmarks:
                landmarks = hand_landmarks[0]
            aperture = gripper_aperture(landmarks) if landmarks else 1.0
            q_rad = aperture * (jm.max_rad - jm.min_rad) + jm.min_rad
            dq_val = (q_rad - prev_dq_cache[i]) / dt if dt > 0 else 0.0
        else:
            angle = joint_angles.get(jm.human_side, {}).get(jm.human_joint, 0.0)
            q_rad = _clamp(math.radians(angle), jm.min_rad, jm.max_rad)
            if prev_angles and dt > 0:
                prev_angle = prev_angles.get(jm.human_side, {}).get(jm.human_joint, angle)
                dq_val = math.radians(angle - prev_angle) / dt
            else:
                dq_val = 0.0

        ddq_val = (dq_val - prev_dq_cache[i]) / dt if dt > 0 else 0.0

        q.append(round(q_rad, 6))
        dq.append(round(dq_val, 6))
        ddq.append(round(ddq_val, 6))

    return {"q": q, "dq": dq, "ddq": ddq}


def build_observation_vector(
    motor_state: dict,
    prev_motor_state: dict | None,
    t: float,
    robot: RobotConfig | None = None,
    period: float = 0.8,
) -> list[float]:
    if robot is None:
        robot = ROBOTS["generic_bimanual"]

    dof = robot.dof
    q   = motor_state.get("q",  [0.0] * dof)
    dq  = motor_state.get("dq", [0.0] * dof)
    q0  = [0.0] * dof

    prev_action = prev_motor_state.get("q", q0) if prev_motor_state else q0

    phase = t % period
    sin_phase = round(math.sin(2 * math.pi * phase / period), 6)
    cos_phase = round(math.cos(2 * math.pi * phase / period), 6)

    return (
        [round(q_i - q0_i, 6) for q_i, q0_i in zip(q, q0)]
        + [round(v, 6) for v in dq]
        + [round(v, 6) for v in prev_action]
        + [sin_phase, cos_phase]
    )
