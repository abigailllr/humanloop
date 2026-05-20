import base64
import json
import os

import cv2

_model = None


def _load():
    global _model
    if _model is None:
        api_key = os.getenv("GEMINI_API_KEY")
        if not api_key:
            return None
        import google.generativeai as genai
        genai.configure(api_key=api_key)
        _model = genai.GenerativeModel(
            "gemini-2.5-flash",
            generation_config={"response_mime_type": "application/json", "temperature": 0.1},
        )
    return _model


def _encode(frame_bgr) -> dict:
    _, buf = cv2.imencode(".jpg", frame_bgr, [cv2.IMWRITE_JPEG_QUALITY, 85])
    return {"mime_type": "image/jpeg", "data": base64.b64encode(buf.tobytes()).decode()}


def validate(video_path: str, challenge_title: str) -> dict:
    model = _load()
    if model is None:
        return {"skipped": True}

    cap = cv2.VideoCapture(video_path)
    total = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    indices = [total // 4, total // 2, 3 * total // 4] if total >= 4 else list(range(total))

    frames = []
    for idx in indices:
        cap.set(cv2.CAP_PROP_POS_FRAMES, idx)
        ok, frame = cap.read()
        if ok:
            frames.append(frame)
    cap.release()

    if not frames:
        return {"skipped": True}

    parts = [
        f'Challenge: "{challenge_title}". Analyze these sampled frames from a submission video. '
        "Return JSON with these keys: "
        "valid (bool, true if a real human is completing the challenge), "
        "confidence (float 0-1), "
        "scene (string describing what is happening), "
        "reason (string explaining the validity decision), "
        "phases (array of {label (one of: approach, grasp, manipulate, transport, place, release), start_pct (float 0-1), end_pct (float 0-1)}), "
        "synthetic (bool, true if the video appears AI-generated or synthetic rather than real camera footage), "
        "synthetic_confidence (float 0-1, how confident you are that this is synthetic), "
        "synthetic_signals (array of strings, each describing a specific indicator of AI generation — "
        "look for: perfect lighting with no natural variation, no motion blur on fast movement, "
        "no film grain or sensor noise, skin texture that looks rendered not photographed, "
        "hands with incorrect anatomy or too many/few fingers, background that looks painted or too clean, "
        "unnaturally stable camera with zero shake during handheld motion, "
        "objects that float or have incorrect shadows, temporal inconsistency between frames).",
        *[_encode(f) for f in frames],
    ]

    try:
        response = model.generate_content(parts)
        return json.loads(response.text)
    except Exception as e:
        return {"skipped": True, "error": str(e)}
