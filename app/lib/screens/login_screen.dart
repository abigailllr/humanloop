import 'package:flutter/material.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../services/auth_service.dart';
import '../main.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  bool _loading = false;

  Future<void> _signIn(Future<dynamic> Function() method) async {
    if (_loading) return;
    setState(() => _loading = true);
    try {
      await method();
      if (mounted) {
        Navigator.pushAndRemoveUntil(
          context,
          MaterialPageRoute(builder: (_) => const RootNav()),
          (_) => false,
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Sign-in failed: $e'), backgroundColor: AppColors.danger),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            children: [
              const Spacer(flex: 2),

              // Logo + wordmark
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

              // Sign-in buttons
              if (_loading)
                const CircularProgressIndicator(color: AppColors.primary)
              else ...[
                _SocialButton(
                  onTap: () => _signIn(AuthService.signInWithGoogle),
                  icon: _GoogleIcon(),
                  label: 'Continue with Google',
                  backgroundColor: Colors.white,
                  foregroundColor: AppColors.textPrimary,
                  borderColor: AppColors.border,
                ),
                const SizedBox(height: 12),
                _SocialButton(
                  onTap: () => _signIn(AuthService.signInWithApple),
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

/// Simple painted Google "G" logo — avoids needing an asset file.
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

    // Draw colored arcs approximating the Google "G"
    final colors = [
      const Color(0xFF4285F4), // blue
      const Color(0xFF34A853), // green
      const Color(0xFFFBBC05), // yellow
      const Color(0xFFEA4335), // red
    ];

    final sweeps = [
      [0.0, 1.6],   // blue — right arc
      [1.6, 1.6],   // green — bottom arc
      [3.2, 0.8],   // yellow — left arc
      [4.0, 2.28],  // red — top arc
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

    // White cutout for the "G" bar
    final whitePaint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.fill;
    canvas.drawRect(Rect.fromLTWH(cx, cy - size.height * 0.1, r * 0.8, size.height * 0.2), whitePaint);
  }

  @override
  bool shouldRepaint(covariant CustomPainter old) => false;
}
