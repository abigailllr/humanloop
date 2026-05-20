package handlers

import (
	"encoding/json"
	"net/http"
)

var supportedRobots = []map[string]any{
	{"id": "generic_bimanual", "name": "Generic Bimanual", "dof": 10, "has_gripper": false},
	{"id": "so100", "name": "SO-ARM100", "dof": 6, "has_gripper": true},
	{"id": "ur5", "name": "Universal Robots UR5", "dof": 6, "has_gripper": false},
	{"id": "franka", "name": "Franka Panda", "dof": 8, "has_gripper": true},
	{"id": "lite6", "name": "UFactory Lite 6", "dof": 7, "has_gripper": true},
}

func GetRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(supportedRobots)
}
