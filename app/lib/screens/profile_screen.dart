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

  @override
  void initState() {
    super.initState();
    final token = ref.read(authProvider).token ?? '';
    _submissions = ApiService.getUserSubmissions(token: token);
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
                      _ActivityCard(submissions: submissions, verified: verified),
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
              const SliverPadding(padding: EdgeInsets.only(bottom: 32)),
            ],
          );
        },
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
  const _ActivityCard({required this.submissions, required this.verified});

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
          _ActivityRow(icon: Icons.bolt_outlined, label: 'Challenges Done', value: '$submissions'),
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
