import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/auth_provider.dart';
import '../services/api_service.dart';
import '../models/submission.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../widgets/submission_card.dart';

class HistoryScreen extends ConsumerStatefulWidget {
  const HistoryScreen({super.key});

  @override
  ConsumerState<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends ConsumerState<HistoryScreen> {
  late Future<List<Submission>> _submissions;

  @override
  void initState() {
    super.initState();
    final token = ref.read(authProvider).token ?? '';
    _submissions = ApiService.getUserSubmissions(token: token);
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: FutureBuilder<List<Submission>>(
        future: _submissions,
        builder: (ctx, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }
          final list = snap.data ?? [];
          final verified = list.where((s) => s.status == SubmissionStatus.verified).length;
          final credits = list.fold(0, (sum, s) => sum + s.creditsEarned);

          return CustomScrollView(
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
                      _SummaryBanner(submissions: list.length, verified: verified, credits: credits),
                      const SizedBox(height: 24),
                      Text('Submissions', style: AppTextStyles.sectionTitle),
                      const SizedBox(height: 4),
                    ],
                  ),
                ),
              ),
              list.isEmpty
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
                          (_, i) => SubmissionCard(submission: list[i]),
                          childCount: list.length,
                        ),
                      ),
                    ),
              const SliverPadding(padding: EdgeInsets.only(bottom: 16)),
            ],
          );
        },
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
