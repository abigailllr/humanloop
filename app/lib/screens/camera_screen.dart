import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/challenge.dart';
import '../providers/auth_provider.dart';
import '../services/api_service.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';

class CameraScreen extends ConsumerStatefulWidget {
  final Challenge? challenge;
  const CameraScreen({super.key, this.challenge});

  @override
  ConsumerState<CameraScreen> createState() => _CameraScreenState();
}

class _CameraScreenState extends ConsumerState<CameraScreen> {
  CameraController? _controller;
  bool _initialized = false;
  bool _recording = false;
  bool _uploading = false;
  bool _submitted = false;
  String? _videoPath;
  String? _error;

  @override
  void initState() {
    super.initState();
    if (widget.challenge != null) _initCamera();
  }

  Future<void> _initCamera() async {
    try {
      final cameras = await availableCameras();
      if (cameras.isEmpty) {
        setState(() => _error = 'No camera found');
        return;
      }
      _controller = CameraController(cameras.first, ResolutionPreset.high, enableAudio: true);
      await _controller!.initialize();
      if (mounted) setState(() => _initialized = true);
    } catch (e) {
      if (mounted) setState(() => _error = e.toString());
    }
  }

  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }

  Future<void> _toggleRecord() async {
    if (_controller == null || !_initialized) return;

    if (_recording) {
      final file = await _controller!.stopVideoRecording();
      setState(() {
        _recording = false;
        _videoPath = file.path;
      });
    } else {
      await _controller!.startVideoRecording();
      setState(() => _recording = true);
    }
  }

  Future<void> _submit() async {
    if (_videoPath == null) return;
    setState(() => _uploading = true);
    final token = ref.read(authProvider).token ?? '';
    final result = await ApiService.uploadVideo(
      challengeId: widget.challenge!.id,
      videoPath: _videoPath!,
      token: token,
    );
    if (mounted) {
      setState(() {
        _uploading = false;
        _submitted = result['submission_id'] != null || result['error'] == null;
      });
    }
  }

  void _retake() => setState(() {
        _videoPath = null;
        _submitted = false;
      });

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
          fit: StackFit.expand,
          children: [
            _buildPreview(),
            _buildBackButton(),
            Positioned(
              bottom: 0,
              left: 0,
              right: 0,
              child: _buildControls(c),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPreview() {
    if (_error != null) {
      return Center(child: Text(_error!, style: const TextStyle(color: Colors.white)));
    }
    if (_videoPath != null) {
      return const ColoredBox(
        color: Color(0xFF111111),
        child: Center(child: Icon(Icons.videocam_rounded, size: 80, color: AppColors.primary)),
      );
    }
    if (!_initialized || _controller == null) {
      return const ColoredBox(
        color: Color(0xFF111111),
        child: Center(child: CircularProgressIndicator(color: AppColors.primary)),
      );
    }
    return CameraPreview(_controller!);
  }

  Widget _buildBackButton() {
    return Positioned(
      top: 16,
      left: 16,
      child: GestureDetector(
        onTap: () => Navigator.pop(context),
        child: Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.15),
            shape: BoxShape.circle,
          ),
          child: const Icon(Icons.arrow_back, color: Colors.white, size: 20),
        ),
      ),
    );
  }

  Widget _buildControls(Challenge c) {
    return Container(
      padding: const EdgeInsets.fromLTRB(24, 24, 24, 40),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.bottomCenter,
          end: Alignment.topCenter,
          colors: [Colors.black, Colors.black.withValues(alpha: 0)],
        ),
      ),
      child: _submitted
          ? _SuccessState(onBack: () => Navigator.pop(context))
          : _videoPath != null
              ? _PreviewControls(uploading: _uploading, onSubmit: _submit, onRetake: _retake)
              : _RecordingControls(challenge: c, recording: _recording, initialized: _initialized, onToggle: _toggleRecord),
    );
  }
}

class _RecordingControls extends StatelessWidget {
  final Challenge challenge;
  final bool recording;
  final bool initialized;
  final VoidCallback onToggle;

  const _RecordingControls({
    required this.challenge,
    required this.recording,
    required this.initialized,
    required this.onToggle,
  });

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
          onTap: initialized ? onToggle : null,
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
          recording ? 'Recording...' : initialized ? 'Tap to record' : 'Initializing...',
          style: TextStyle(color: Colors.white.withValues(alpha: 0.7), fontSize: 13),
        ),
      ],
    );
  }
}

class _PreviewControls extends StatelessWidget {
  final bool uploading;
  final VoidCallback onSubmit;
  final VoidCallback onRetake;

  const _PreviewControls({required this.uploading, required this.onSubmit, required this.onRetake});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        const Icon(Icons.check_circle_outline, color: Colors.white, size: 48),
        const SizedBox(height: 12),
        const Text('Video recorded', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: Colors.white)),
        const SizedBox(height: 24),
        if (uploading)
          const CircularProgressIndicator(color: Colors.white)
        else ...[
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: onSubmit,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                elevation: 0,
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
              ),
              child: Text('Submit', style: AppTextStyles.buttonLabel),
            ),
          ),
          const SizedBox(height: 10),
          TextButton(
            onPressed: onRetake,
            child: const Text('Retake', style: TextStyle(color: Colors.white70, fontSize: 14)),
          ),
        ],
      ],
    );
  }
}

class _SuccessState extends StatelessWidget {
  final VoidCallback onBack;
  const _SuccessState({required this.onBack});

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
            onPressed: onBack,
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
