import json
import math
import subprocess


_KNOWN_AI_ENCODERS = {
    "sora", "runway", "pika", "kling", "gen-2", "gen-3",
    "stable-video", "animatediff", "deforum", "modelscope",
}


def check_metadata(video_path: str) -> dict:
    try:
        result = subprocess.run(
            [
                "ffprobe", "-v", "quiet", "-print_format", "json",
                "-show_format", "-show_streams", video_path,
            ],
            capture_output=True, text=True, timeout=15,
        )
        info = json.loads(result.stdout)
    except Exception as e:
        return {"error": str(e), "suspicious": False}

    fmt = info.get("format", {})
    tags = {k.lower(): v.lower() for k, v in fmt.get("tags", {}).items()}
    signals = []

    encoder = tags.get("encoder", "") or tags.get("software", "") or tags.get("creation_tool", "")
    for ai_tool in _KNOWN_AI_ENCODERS:
        if ai_tool in encoder:
            signals.append(f"ai_encoder:{ai_tool}")

    missing = []
    for field in ("make", "model", "com.apple.quicktime.model", "android.version"):
        if field not in tags:
            missing.append(field)
    if len(missing) == 4:
        signals.append("no_device_metadata")

    if "creation_time" not in tags and "date" not in tags:
        signals.append("no_creation_time")

    streams = info.get("streams", [])
    video_streams = [s for s in streams if s.get("codec_type") == "video"]
    if video_streams:
        vstream = video_streams[0]
        codec = vstream.get("codec_name", "")
        if codec not in ("h264", "hevc", "h265", "vp9", "av1", "prores", "mjpeg"):
            signals.append(f"unusual_codec:{codec}")

        r_frame_rate = vstream.get("r_frame_rate", "0/1")
        try:
            num, den = r_frame_rate.split("/")
            fps = int(num) / int(den)
            if fps in (24.0, 25.0, 30.0, 60.0) and fps == round(fps):
                signals.append("suspiciously_round_fps")
        except Exception:
            pass

    return {
        "suspicious": len(signals) > 0,
        "signals": signals,
        "encoder": encoder,
    }


def check_motion_naturalness(frames: list) -> dict:
    pose_sequences = []
    for frame in frames:
        pose = frame.get("pose")
        if pose and len(pose) >= 11:
            wrist_l = pose[15]
            wrist_r = pose[16]
            pose_sequences.append((wrist_l["x"], wrist_l["y"], wrist_r["x"], wrist_r["y"]))

    if len(pose_sequences) < 10:
        return {"skipped": True}

    diffs = []
    for i in range(1, len(pose_sequences)):
        prev, curr = pose_sequences[i - 1], pose_sequences[i]
        delta = math.sqrt(sum((c - p) ** 2 for c, p in zip(curr, prev)))
        diffs.append(delta)

    mean = sum(diffs) / len(diffs)
    variance = sum((d - mean) ** 2 for d in diffs) / len(diffs)
    std = math.sqrt(variance)

    signals = []

    if mean < 0.001:
        signals.append("near_zero_motion")

    if mean > 0.001 and std / (mean + 1e-9) < 0.05:
        signals.append("unnaturally_smooth_motion")

    jerk = [abs(diffs[i] - diffs[i - 1]) for i in range(1, len(diffs))]
    mean_jerk = sum(jerk) / len(jerk) if jerk else 0
    if mean_jerk < 1e-5 and mean > 0.001:
        signals.append("zero_jerk")

    return {
        "suspicious": len(signals) > 0,
        "signals": signals,
        "motion_mean": round(mean, 6),
        "motion_std": round(std, 6),
    }
