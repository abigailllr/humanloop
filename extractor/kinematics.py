import math


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

        trunk_axis  = _vec(hip, shoulder)
        lateral     = _vec(shoulder, opp_sh) if side == "l" else _vec(opp_sh, shoulder)
        upper_arm   = _vec(shoulder, elbow)
        forearm     = _vec(elbow, wrist)

        shoulder_flex  = _signed_angle_deg(trunk_axis, upper_arm, lateral)
        shoulder_abduct = _angle_deg(lateral, upper_arm)
        elbow_flex     = 180.0 - _angle_deg(upper_arm, forearm)

        index  = lm(f"{side}_index")
        pinky  = lm(f"{side}_pinky")
        hand   = _vec(wrist, index)
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
