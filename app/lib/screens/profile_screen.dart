import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/auth_provider.dart';
import '../services/api_service.dart';
import '../models/submission.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../widgets/submission_card.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  const ProfileScreen({super.key});

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen> {
  late Future<List<Submission>> _submissions;
  late Future<Map<String, dynamic>> _stats;
  late Future<List<Map<String, dynamic>>> _creditHistory;

  @override
  void initState() {
    super.initState();
    final token = ref.read(authProvider).token ?? '';
    _submissions = ApiService.getUserSubmissions(token: token);
    _stats = ApiService.getUserStats(token: token);
    _creditHistory = ApiService.getCreditHistory(token: token);
  }

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authProvider).user;
    final credits = user?.credits ?? 0;
    final submissions = user?.submissions ?? 0;
    final name = user?.name ?? 'Contributor';

    return SafeArea(
      child: FutureBuilder<List<Submission>>(
        future: _submissions,
        builder: (ctx, snap) {
          final list = snap.data ?? [];
          final verified = list.where((s) => s.status == SubmissionStatus.verified).length;
          final earned = list.fold(0, (sum, s) => sum + s.creditsEarned);

          return CustomScrollView(
            slivers: [
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(24, 24, 24, 0),
                sliver: SliverToBoxAdapter(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Profile', style: AppTextStyles.screenTitle),
                      const SizedBox(height: 28),
                      _AvatarCard(name: name, submissions: submissions),
                      const SizedBox(height: 20),
                      _BalanceCard(credits: credits),
                      const SizedBox(height: 20),
                      FutureBuilder<Map<String, dynamic>>(
                        future: _stats,
                        builder: (_, ss) {
                          final qualityAvg = (ss.data?['quality_avg'] as num?)?.toDouble() ?? 0.0;
                          return _ActivityCard(submissions: submissions, verified: verified, qualityAvg: qualityAvg);
                        },
                      ),
                      const SizedBox(height: 28),
                      SizedBox(
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: () => ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Cash out coming soon.')),
                          ),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: AppColors.primary,
                            foregroundColor: Colors.white,
                            elevation: 0,
                            padding: const EdgeInsets.symmetric(vertical: 16),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                          ),
                          child: Text('Cash Out', style: AppTextStyles.buttonLabel),
                        ),
                      ),
                      const SizedBox(height: 32),
                      if (list.isNotEmpty || snap.connectionState == ConnectionState.done) ...[
                        Text('My Submissions', style: AppTextStyles.sectionTitle),
                        const SizedBox(height: 4),
                        if (list.isNotEmpty) ...[
                          Text('${list.length} total · $verified verified · $earned credits', style: AppTextStyles.caption),
                          const SizedBox(height: 16),
                        ],
                      ],
                    ],
                  ),
                ),
              ),
              if (snap.connectionState == ConnectionState.waiting)
                const SliverToBoxAdapter(
                  child: Center(child: Padding(
                    padding: EdgeInsets.only(top: 24),
                    child: CircularProgressIndicator(),
                  )),
                )
              else if (list.isEmpty)
                SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.only(top: 24),
                    child: Column(
                      children: [
                        Icon(Icons.videocam_off_outlined, size: 40, color: AppColors.textTertiary),
                        const SizedBox(height: 10),
                        Text('No submissions yet', style: AppTextStyles.bodyMedium),
                      ],
                    ),
                  ),
                )
              else
                SliverPadding(
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  sliver: SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (_, i) => SubmissionCard(submission: list[i]),
                      childCount: list.length,
                    ),
                  ),
                ),
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(24, 32, 24, 0),
                sliver: SliverToBoxAdapter(
                  child: FutureBuilder<List<Map<String, dynamic>>>(
                    future: _creditHistory,
                    builder: (_, hs) {
                      final history = hs.data ?? [];
                      if (hs.connectionState == ConnectionState.waiting || history.isEmpty) {
                        return const SizedBox.shrink();
                      }
                      return Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Credit History', style: AppTextStyles.sectionTitle),
                          const SizedBox(height: 16),
                          ...history.map((e) => _CreditHistoryRow(entry: e)),
                        ],
                      );
                    },
                  ),
                ),
              ),
              const SliverPadding(padding: EdgeInsets.only(bottom: 32)),
            ],
          );
        },
      ),
    );
  }
}

class _CreditHistoryRow extends StatelessWidget {
  final Map<String, dynamic> entry;
  const _CreditHistoryRow({required this.entry});

  @override
  Widget build(BuildContext context) {
    final amount = entry['amount'] as int? ?? 0;
    final reason = entry['reason'] as String? ?? '';
    final createdAt = entry['created_at'] as String? ?? '';
    final positive = amount > 0;
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.border),
      ),
      child: Row(
        children: [
          Container(
            width: 36,
            height: 36,
            decoration: BoxDecoration(
              color: positive ? AppColors.successLight : AppColors.dangerLight,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(positive ? Icons.add_rounded : Icons.remove_rounded, size: 18, color: positive ? AppColors.success : AppColors.danger),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(reason, style: AppTextStyles.bodyMedium),
                if (createdAt.isNotEmpty) Text(createdAt.substring(0, 10), style: AppTextStyles.caption),
              ],
            ),
          ),
          Text(
            '${positive ? '+' : ''}$amount cr',
            style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: positive ? AppColors.success : AppColors.danger),
          ),
        ],
      ),
    );
  }
}

class _AvatarCard extends StatelessWidget {
  final String name;
  final int submissions;
  const _AvatarCard({required this.name, required this.submissions});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.surfaceGray,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        children: [
          Container(
            width: 60,
            height: 60,
            decoration: const BoxDecoration(color: AppColors.primary, shape: BoxShape.circle),
            child: const Icon(Icons.person, color: Colors.white, size: 32),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(name, style: AppTextStyles.cardTitle),
                const SizedBox(height: 2),
                Text('$submissions videos filmed', style: AppTextStyles.bodySmall),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
            decoration: BoxDecoration(
              color: AppColors.success.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text('Active', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.success)),
          ),
        ],
      ),
    );
  }
}

class _BalanceCard extends StatelessWidget {
  final int credits;
  const _BalanceCard({required this.credits});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: AppColors.primary,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Available Credits', style: TextStyle(fontSize: 13, color: Colors.white70, fontWeight: FontWeight.w500)),
          const SizedBox(height: 8),
          Text('$credits', style: const TextStyle(fontSize: 48, fontWeight: FontWeight.w800, color: Colors.white, letterSpacing: -1)),
          const SizedBox(height: 4),
          const Text('credits earned', style: TextStyle(fontSize: 13, color: Colors.white60)),
        ],
      ),
    );
  }
}

class _ActivityCard extends StatelessWidget {
  final int submissions;
  final int verified;
  final double qualityAvg;
  const _ActivityCard({required this.submissions, required this.verified, required this.qualityAvg});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
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
          Text('Your Activity', style: AppTextStyles.cardTitle.copyWith(fontSize: 15)),
          const SizedBox(height: 16),
          _ActivityRow(icon: Icons.videocam_outlined, label: 'Videos Submitted', value: '$submissions'),
          const SizedBox(height: 14),
          _ActivityRow(icon: Icons.check_circle_outline, label: 'Verified', value: '$verified'),
          const SizedBox(height: 14),
          _ActivityRow(icon: Icons.star_outline_rounded, label: 'Avg Quality', value: qualityAvg > 0 ? '${(qualityAvg * 100).round()}%' : '—'),
        ],
      ),
    );
  }
}

class _ActivityRow extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  const _ActivityRow({required this.icon, required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(color: AppColors.primaryLight, borderRadius: BorderRadius.circular(10)),
          child: Icon(icon, size: 18, color: AppColors.primary),
        ),
        const SizedBox(width: 12),
        Expanded(child: Text(label, style: AppTextStyles.bodyMedium)),
        Text(value, style: AppTextStyles.cardTitle.copyWith(fontSize: 15)),
      ],
    );
  }
}
