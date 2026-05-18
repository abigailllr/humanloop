import 'package:flutter/material.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../models/submission.dart';
import '../widgets/submission_card.dart';

final _demoSubmissions = [
  Submission(
    id: 's1',
    challengeId: 'c4',
    challengeTitle: 'Pour & Fill',
    submittedAt: DateTime.now().subtract(const Duration(hours: 2)),
    status: SubmissionStatus.verified,
    creditsEarned: 2,
  ),
  Submission(
    id: 's2',
    challengeId: 'c1',
    challengeTitle: 'Pick & Place',
    submittedAt: DateTime.now().subtract(const Duration(days: 1)),
    status: SubmissionStatus.verified,
    creditsEarned: 1,
  ),
  Submission(
    id: 's3',
    challengeId: 'c2',
    challengeTitle: 'Fold It',
    submittedAt: DateTime.now().subtract(const Duration(days: 3)),
    status: SubmissionStatus.pending,
    creditsEarned: 0,
  ),
];

class HistoryScreen extends StatelessWidget {
  const HistoryScreen({super.key});

  int get _totalCredits =>
      _demoSubmissions.fold(0, (sum, s) => sum + s.creditsEarned);

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: CustomScrollView(
        slivers: [
          SliverPadding(
            padding: const EdgeInsets.fromLTRB(24, 24, 24, 8),
            sliver: SliverToBoxAdapter(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('History', style: AppTextStyles.screenTitle),
                  const SizedBox(height: 4),
                  Text('Your past submissions', style: AppTextStyles.bodySmall),
                  const SizedBox(height: 20),
                  _SummaryBanner(
                    submissions: _demoSubmissions.length,
                    verified: _demoSubmissions.where((s) => s.status == SubmissionStatus.verified).length,
                    credits: _totalCredits,
                  ),
                  const SizedBox(height: 24),
                  Text('Submissions', style: AppTextStyles.sectionTitle),
                  const SizedBox(height: 4),
                ],
              ),
            ),
          ),
          _demoSubmissions.isEmpty
              ? SliverFillRemaining(
                  child: Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.videocam_off_outlined, size: 48, color: AppColors.textTertiary),
                        const SizedBox(height: 12),
                        Text('No submissions yet', style: AppTextStyles.bodyMedium),
                      ],
                    ),
                  ),
                )
              : SliverPadding(
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  sliver: SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (_, i) => SubmissionCard(submission: _demoSubmissions[i]),
                      childCount: _demoSubmissions.length,
                    ),
                  ),
                ),
          const SliverPadding(padding: EdgeInsets.only(bottom: 16)),
        ],
      ),
    );
  }
}

class _SummaryBanner extends StatelessWidget {
  final int submissions;
  final int verified;
  final int credits;

  const _SummaryBanner({required this.submissions, required this.verified, required this.credits});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.primary,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        children: [
          _BannerStat(value: '$submissions', label: 'Total'),
          _divider(),
          _BannerStat(value: '$verified', label: 'Verified'),
          _divider(),
          _BannerStat(value: '$credits', label: 'Credits'),
        ],
      ),
    );
  }

  Widget _divider() => Container(width: 1, height: 36, color: Colors.white24, margin: const EdgeInsets.symmetric(horizontal: 16));
}

class _BannerStat extends StatelessWidget {
  final String value;
  final String label;
  const _BannerStat({required this.value, required this.label});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Column(
        children: [
          Text(value, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w800, color: Colors.white, letterSpacing: -0.5)),
          const SizedBox(height: 2),
          Text(label, style: const TextStyle(fontSize: 12, color: Colors.white60, fontWeight: FontWeight.w500)),
        ],
      ),
    );
  }
}
