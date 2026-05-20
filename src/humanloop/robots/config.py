import math
from dataclasses import dataclass, field


@dataclass
class JointMapping:
    robot_joint: str
    human_side: str
    human_joint: str
    min_rad: float
    max_rad: float
    const_val: float = 0.0


@dataclass
class RobotConfig:
    name: str
    robot_type: str
    joints: list[JointMapping]

    @property
    def dof(self) -> int:
        return len(self.joints)

    @property
    def has_gripper(self) -> bool:
        return any(j.human_side == "gripper" for j in self.joints)

    @property
    def obs_dim(self) -> int:
        return self.dof * 3 + 2

    @property
    def action_dim(self) -> int:
        return self.dof


def _r(deg: float) -> float:
    return round(math.radians(deg), 6)


ROBOTS: dict[str, RobotConfig] = {
    "generic_bimanual": RobotConfig(
        name="generic_bimanual",
        robot_type="generic_bimanual_arm",
        joints=[
            JointMapping("left_shoulder_flexion",   "left",  "shoulder_flexion",   _r(-180), _r(180)),
            JointMapping("left_shoulder_abduction",  "left",  "shoulder_abduction", _r(0),    _r(180)),
            JointMapping("left_elbow_flexion",       "left",  "elbow_flexion",      _r(0),    _r(145)),
            JointMapping("left_wrist_flexion",       "left",  "wrist_flexion",      _r(-80),  _r(80)),
            JointMapping("left_wrist_deviation",     "left",  "wrist_deviation",    _r(-40),  _r(40)),
            JointMapping("right_shoulder_flexion",   "right", "shoulder_flexion",   _r(-180), _r(180)),
            JointMapping("right_shoulder_abduction", "right", "shoulder_abduction", _r(0),    _r(180)),
            JointMapping("right_elbow_flexion",      "right", "elbow_flexion",      _r(0),    _r(145)),
            JointMapping("right_wrist_flexion",      "right", "wrist_flexion",      _r(-80),  _r(80)),
            JointMapping("right_wrist_deviation",    "right", "wrist_deviation",    _r(-40),  _r(40)),
        ],
    ),
    "so100": RobotConfig(
        name="so100",
        robot_type="so_arm100",
        joints=[
            JointMapping("shoulder_pan",  "right", "shoulder_abduction", _r(-180), _r(180)),
            JointMapping("shoulder_lift", "right", "shoulder_flexion",   _r(-90),  _r(90)),
            JointMapping("elbow_flex",    "right", "elbow_flexion",      _r(-90),  _r(90)),
            JointMapping("wrist_flex",    "right", "wrist_flexion",      _r(-90),  _r(90)),
            JointMapping("wrist_roll",    "right", "wrist_deviation",    _r(-90),  _r(90)),
            JointMapping("gripper",       "gripper", "",                  0.0,      1.0),
        ],
    ),
    "ur5": RobotConfig(
        name="ur5",
        robot_type="ur5",
        joints=[
            JointMapping("base",      "right", "shoulder_abduction", _r(-360), _r(360)),
            JointMapping("shoulder",  "right", "shoulder_flexion",   _r(-360), _r(360)),
            JointMapping("elbow",     "right", "elbow_flexion",      _r(-360), _r(360)),
            JointMapping("wrist_1",   "right", "wrist_flexion",      _r(-360), _r(360)),
            JointMapping("wrist_2",   "right", "wrist_deviation",    _r(-360), _r(360)),
            JointMapping("wrist_3",   "const", "",                   _r(-360), _r(360)),
        ],
    ),
    "franka": RobotConfig(
        name="franka",
        robot_type="franka_panda",
        joints=[
            JointMapping("panda_joint1", "right", "shoulder_abduction", _r(-166), _r(166)),
            JointMapping("panda_joint2", "right", "shoulder_flexion",   _r(-101), _r(101)),
            JointMapping("panda_joint3", "const", "",                   _r(-166), _r(166)),
            JointMapping("panda_joint4", "right", "elbow_flexion",      _r(-176), _r(-4)),
            JointMapping("panda_joint5", "right", "wrist_flexion",      _r(-166), _r(166)),
            JointMapping("panda_joint6", "right", "wrist_deviation",    _r(-1),   _r(215)),
            JointMapping("panda_joint7", "const", "",                   _r(-166), _r(166)),
            JointMapping("gripper",      "gripper", "",                  0.0,      1.0),
        ],
    ),
    "lite6": RobotConfig(
        name="lite6",
        robot_type="ufactory_lite6",
        joints=[
            JointMapping("joint1", "right", "shoulder_abduction", _r(-360), _r(360)),
            JointMapping("joint2", "right", "shoulder_flexion",   _r(-130), _r(130)),
            JointMapping("joint3", "right", "elbow_flexion",      _r(-300), _r(60)),
            JointMapping("joint4", "right", "wrist_flexion",      _r(-360), _r(360)),
            JointMapping("joint5", "right", "wrist_deviation",    _r(-124), _r(124)),
            JointMapping("joint6", "const", "",                   _r(-360), _r(360)),
            JointMapping("gripper", "gripper", "",                 0.0,      1.0),
        ],
    ),
}


def get_robot(name: str) -> RobotConfig:
    if name not in ROBOTS:
        raise ValueError(f"unknown robot '{name}', available: {list(ROBOTS)}")
    return ROBOTS[name]
