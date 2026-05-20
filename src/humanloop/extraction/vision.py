from ultralytics import YOLO
import numpy as np

_model = None


def _load():
    global _model
    if _model is None:
        _model = YOLO("yolo11x.pt")
    return _model


def detect(frame_rgb: np.ndarray) -> list[dict]:
    model = _load()
    results = model(frame_rgb, verbose=False)
    objects = []
    for box in results[0].boxes:
        cx, cy, w, h = box.xywhn[0].tolist()
        objects.append({
            "label": results[0].names[int(box.cls)],
            "confidence": round(float(box.conf), 3),
            "bbox": [round(cx, 4), round(cy, 4), round(w, 4), round(h, 4)],
        })
    return objects


def hand_contacts(hands: list[list[dict]], objects: list[dict]) -> list[dict]:
    contacts = []
    for hand_idx, hand in enumerate(hands):
        if not hand:
            continue
        wrist = hand[0]
        tip_indices = [4, 8, 12, 16, 20]
        tips = [hand[i] for i in tip_indices if i < len(hand)]
        points = [wrist] + tips

        for obj in objects:
            cx, cy, w, h = obj["bbox"]
            x1, y1 = cx - w / 2, cy - h / 2
            x2, y2 = cx + w / 2, cy + h / 2

            hits = sum(1 for p in points if x1 <= p["x"] <= x2 and y1 <= p["y"] <= y2)
            confidence = round(hits / len(points), 2)

            if confidence > 0.2:
                contacts.append({
                    "hand": hand_idx,
                    "object_label": obj["label"],
                    "confidence": confidence,
                })

    return contacts
