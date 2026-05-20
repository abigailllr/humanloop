import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../providers/auth_provider.dart';
import '../models/user.dart';
import '../main.dart';

class LoginScreen extends ConsumerWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auth = ref.watch(authProvider);

    Future<void> signIn(Future<void> Function() action) async {
      try {
        await action();
        if (context.mounted) {
          Navigator.pushAndRemoveUntil(
            context,
            MaterialPageRoute(builder: (_) => const RootNav()),
            (_) => false,
          );
        }
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Sign-in failed: $e'), backgroundColor: AppColors.danger),
          );
        }
      }
    }

    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            children: [
              const Spacer(flex: 2),
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(color: AppColors.primary, borderRadius: BorderRadius.circular(24)),
                child: const Icon(Icons.precision_manufacturing_rounded, color: Colors.white, size: 44),
              ),
              const SizedBox(height: 24),
              Text('HumanLoop', style: AppTextStyles.screenTitle.copyWith(fontSize: 32)),
              const SizedBox(height: 8),
              Text(
                'Film challenges. Train robots.\nEarn credits.',
                style: AppTextStyles.bodyMedium.copyWith(height: 1.5),
                textAlign: TextAlign.center,
              ),
              const Spacer(flex: 2),
              if (auth.loading)
                const CircularProgressIndicator(color: AppColors.primary)
              else ...[
                _SocialButton(
                  onTap: () => signIn(ref.read(authProvider.notifier).signInWithGoogle),
                  icon: _GoogleIcon(),
                  label: 'Continue with Google',
                  backgroundColor: Colors.white,
                  foregroundColor: AppColors.textPrimary,
                  borderColor: AppColors.border,
                ),
                const SizedBox(height: 12),
                _SocialButton(
                  onTap: () => signIn(ref.read(authProvider.notifier).signInWithApple),
                  icon: const Icon(Icons.apple, color: Colors.white, size: 22),
                  label: 'Continue with Apple',
                  backgroundColor: AppColors.textPrimary,
                  foregroundColor: Colors.white,
                ),
              ],
              const SizedBox(height: 32),
              Text(
                'By continuing, you agree to our Terms of Service\nand Privacy Policy.',
                style: AppTextStyles.caption.copyWith(height: 1.5),
                textAlign: TextAlign.center,
              ),
              if (kDebugMode) ...[
                const SizedBox(height: 16),
                TextButton(
                  onPressed: () {
                    ref.read(authProvider.notifier).devBypass();
                    Navigator.pushAndRemoveUntil(
                      context,
                      MaterialPageRoute(builder: (_) => const RootNav()),
                      (_) => false,
                    );
                  },
                  child: Text('Dev bypass', style: TextStyle(color: AppColors.textSecondary, fontSize: 12)),
                ),
              ],
              const SizedBox(height: 24),
            ],
          ),
        ),
      ),
    );
  }
}

class _SocialButton extends StatelessWidget {
  final VoidCallback onTap;
  final Widget icon;
  final String label;
  final Color backgroundColor;
  final Color foregroundColor;
  final Color? borderColor;

  const _SocialButton({
    required this.onTap,
    required this.icon,
    required this.label,
    required this.backgroundColor,
    required this.foregroundColor,
    this.borderColor,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: OutlinedButton(
        onPressed: onTap,
        style: OutlinedButton.styleFrom(
          backgroundColor: backgroundColor,
          foregroundColor: foregroundColor,
          side: BorderSide(color: borderColor ?? Colors.transparent, width: borderColor != null ? 1 : 0),
          padding: const EdgeInsets.symmetric(vertical: 15),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          elevation: 0,
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            icon,
            const SizedBox(width: 12),
            Text(label, style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: foregroundColor)),
          ],
        ),
      ),
    );
  }
}

class _GoogleIcon extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 22,
      height: 22,
      child: CustomPaint(painter: _GooglePainter()),
    );
  }
}

class _GooglePainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final cx = size.width / 2;
    final cy = size.height / 2;
    final r = size.width / 2;

    final colors = [
      const Color(0xFF4285F4),
      const Color(0xFF34A853),
      const Color(0xFFFBBC05),
      const Color(0xFFEA4335),
    ];

    final sweeps = [
      [0.0, 1.6],
      [1.6, 1.6],
      [3.2, 0.8],
      [4.0, 2.28],
    ];

    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = size.width * 0.2;

    for (int i = 0; i < 4; i++) {
      paint.color = colors[i];
      canvas.drawArc(
        Rect.fromCircle(center: Offset(cx, cy), radius: r * 0.75),
        sweeps[i][0],
        sweeps[i][1],
        false,
        paint,
      );
    }

    final whitePaint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.fill;
    canvas.drawRect(Rect.fromLTWH(cx, cy - size.height * 0.1, r * 0.8, size.height * 0.2), whitePaint);
  }

  @override
  bool shouldRepaint(covariant CustomPainter old) => false;
}
