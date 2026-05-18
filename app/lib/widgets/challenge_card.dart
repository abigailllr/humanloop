import 'package:flutter/material.dart';
import '../models/challenge.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';

class ChallengeCard extends StatelessWidget {
  final Challenge challenge;
  final VoidCallback onTap;

  const ChallengeCard({super.key, required this.challenge, required this.onTap});

  Color get _diffColor {
    switch (challenge.difficulty) {
      case 'Easy':   return AppColors.success;
      case 'Medium': return AppColors.warning;
      default:       return AppColors.danger;
    }
  }

  Color get _diffBg {
    switch (challenge.difficulty) {
      case 'Easy':   return AppColors.successLight;
      case 'Medium': return AppColors.warningLight;
      default:       return AppColors.dangerLight;
    }
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.only(bottom: 14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.border),
          boxShadow: [
            BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 4)),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              height: 140,
              decoration: const BoxDecoration(
                color: AppColors.primaryLight,
                borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
              ),
              child: const Center(
                child: Icon(Icons.videocam_outlined, size: 48, color: AppColors.primary),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                        decoration: BoxDecoration(
                          color: _diffBg,
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: Text(challenge.difficulty, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: _diffColor)),
                      ),
                      const Spacer(),
                      Icon(Icons.people_outline, size: 14, color: Colors.grey.shade400),
                      const SizedBox(width: 4),
                      Text('${challenge.submissions}', style: AppTextStyles.label.copyWith(color: Colors.grey.shade500)),
                    ],
                  ),
                  const SizedBox(height: 10),
                  Text(challenge.title, style: AppTextStyles.cardTitle),
                  const SizedBox(height: 4),
                  Text(challenge.description, style: AppTextStyles.bodySmall),
                  const SizedBox(height: 14),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: onTap,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        foregroundColor: Colors.white,
                        elevation: 0,
                        padding: const EdgeInsets.symmetric(vertical: 13),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: const Text('Film this challenge', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
