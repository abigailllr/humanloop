import 'package:flutter/material.dart';
import '../models/submission.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';

class SubmissionCard extends StatelessWidget {
  final Submission submission;
  const SubmissionCard({super.key, required this.submission});

  Color get _statusColor {
    switch (submission.status) {
      case SubmissionStatus.verified:  return AppColors.success;
      case SubmissionStatus.rejected:  return AppColors.danger;
      case SubmissionStatus.synthetic: return AppColors.danger;
      case SubmissionStatus.pending:   return AppColors.warning;
    }
  }

  Color get _statusBg {
    switch (submission.status) {
      case SubmissionStatus.verified:  return AppColors.successLight;
      case SubmissionStatus.rejected:  return AppColors.dangerLight;
      case SubmissionStatus.synthetic: return AppColors.dangerLight;
      case SubmissionStatus.pending:   return AppColors.warningLight;
    }
  }

  String get _statusLabel {
    switch (submission.status) {
      case SubmissionStatus.verified:  return 'Verified';
      case SubmissionStatus.rejected:  return 'Rejected';
      case SubmissionStatus.synthetic: return 'AI Detected';
      case SubmissionStatus.pending:   return 'Pending';
    }
  }

  String _formatDate(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    return '${diff.inDays}d ago';
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.border),
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 4)),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(color: AppColors.primaryLight, borderRadius: BorderRadius.circular(12)),
            child: const Icon(Icons.videocam_outlined, size: 24, color: AppColors.primary),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(submission.challengeTitle, style: AppTextStyles.cardTitle.copyWith(fontSize: 15)),
                const SizedBox(height: 3),
                Text(_formatDate(submission.submittedAt), style: AppTextStyles.caption),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: _statusBg,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(_statusLabel, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: _statusColor)),
              ),
              if (submission.creditsEarned > 0) ...[
                const SizedBox(height: 4),
                Text('+${submission.creditsEarned} cr', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w700, color: AppColors.success)),
              ],
              if (submission.status == SubmissionStatus.verified && submission.qualityScore > 0) ...[
                const SizedBox(height: 2),
                Text('${(submission.qualityScore * 100).round()}% quality', style: TextStyle(fontSize: 11, color: AppColors.textTertiary)),
              ],
            ],
          ),
        ],
      ),
    );
  }
}
