import 'package:flutter/material.dart';
import '../models/challenge.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';

class CameraScreen extends StatefulWidget {
  final Challenge? challenge;
  const CameraScreen({super.key, this.challenge});

  @override
  State<CameraScreen> createState() => _CameraScreenState();
}

class _CameraScreenState extends State<CameraScreen> {
  bool _recording = false;
  bool _submitted = false;

  void _toggleRecord() {
    setState(() => _recording = !_recording);
    if (_recording) {
      Future.delayed(const Duration(seconds: 3), () {
        if (mounted) setState(() { _recording = false; _submitted = true; });
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = widget.challenge;

    if (c == null) {
      return SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Film', style: AppTextStyles.screenTitle),
              const SizedBox(height: 8),
              Text('Pick a challenge from the feed to start filming.', style: AppTextStyles.bodyMedium),
            ],
          ),
        ),
      );
    }

    return Scaffold(
      backgroundColor: Colors.black,
      body: SafeArea(
        child: Stack(
          children: [
            const SizedBox.expand(
              child: ColoredBox(
                color: Color(0xFF111111),
                child: Center(
                  child: Icon(Icons.videocam, size: 80, color: Color(0xFF333333)),
                ),
              ),
            ),
            Positioned(
              top: 16,
              left: 16,
              child: GestureDetector(
                onTap: () => Navigator.pop(context),
                child: Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.15), shape: BoxShape.circle),
                  child: const Icon(Icons.arrow_back, color: Colors.white, size: 20),
                ),
              ),
            ),
            Positioned(
              bottom: 0,
              left: 0,
              right: 0,
              child: Container(
                padding: const EdgeInsets.fromLTRB(24, 24, 24, 40),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                    colors: [Colors.black, Colors.black.withValues(alpha: 0)],
                  ),
                ),
                child: _submitted ? const _SuccessState() : _RecordingControls(challenge: c, recording: _recording, onToggle: _toggleRecord),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _RecordingControls extends StatelessWidget {
  final Challenge challenge;
  final bool recording;
  final VoidCallback onToggle;

  const _RecordingControls({required this.challenge, required this.recording, required this.onToggle});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(challenge.title, style: AppTextStyles.cardTitle.copyWith(color: Colors.white)),
              const SizedBox(height: 4),
              Text(challenge.description, style: AppTextStyles.bodySmall.copyWith(color: Colors.white.withValues(alpha: 0.7))),
            ],
          ),
        ),
        const SizedBox(height: 24),
        GestureDetector(
          onTap: onToggle,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            width: 72,
            height: 72,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: recording ? AppColors.danger : Colors.white,
              border: Border.all(color: Colors.white.withValues(alpha: 0.4), width: 4),
            ),
            child: recording ? const Icon(Icons.stop, color: Colors.white, size: 30) : const SizedBox(),
          ),
        ),
        const SizedBox(height: 12),
        Text(
          recording ? 'Recording...' : 'Tap to record',
          style: TextStyle(color: Colors.white.withValues(alpha: 0.7), fontSize: 13),
        ),
      ],
    );
  }
}

class _SuccessState extends StatelessWidget {
  const _SuccessState();

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Container(
          width: 64,
          height: 64,
          decoration: const BoxDecoration(color: AppColors.success, shape: BoxShape.circle),
          child: const Icon(Icons.check, color: Colors.white, size: 32),
        ),
        const SizedBox(height: 16),
        const Text('Submitted!', style: TextStyle(fontSize: 22, fontWeight: FontWeight.w800, color: Colors.white)),
        const SizedBox(height: 6),
        Text('Your video is being processed.', style: TextStyle(fontSize: 14, color: Colors.white.withValues(alpha: 0.6))),
        const SizedBox(height: 24),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: () => Navigator.pop(context),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.white,
              foregroundColor: AppColors.textPrimary,
              elevation: 0,
              padding: const EdgeInsets.symmetric(vertical: 14),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
            ),
            child: Text('Back to challenges', style: AppTextStyles.buttonLabel.copyWith(fontSize: 15)),
          ),
        ),
      ],
    );
  }
}
