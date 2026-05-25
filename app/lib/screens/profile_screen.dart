import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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
    final level = user?.level ?? 'Rookie';
    final badges = user?.badges ?? [];
    final referralCode = user?.referralCode ?? '';

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
                      _AvatarCard(name: name, submissions: submissions, level: level),
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
                      if (badges.isNotEmpty) ...[
                        const SizedBox(height: 20),
                        _BadgesCard(badges: badges),
                      ],
                      const SizedBox(height: 20),
                      _ReferralCard(code: referralCode, token: ref.read(authProvider).token ?? ''),
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
  final String level;
  const _AvatarCard({required this.name, required this.submissions, required this.level});

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
              color: AppColors.primary.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(level, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.primary)),
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

const _kBadgeLabels = {
  'first_submission': ('First Step', Icons.flag_rounded),
  'first_verified': ('First Verified', Icons.verified_rounded),
  'ten_verified': ('10 Verified', Icons.workspace_premium_rounded),
  'fifty_verified': ('50 Verified', Icons.emoji_events_rounded),
  'century': ('Century', Icons.military_tech_rounded),
  'quality_star': ('Quality Star', Icons.star_rounded),
};

class _BadgesCard extends StatelessWidget {
  final List<String> badges;
  const _BadgesCard({required this.badges});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.border),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 4))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Badges', style: AppTextStyles.cardTitle.copyWith(fontSize: 15)),
          const SizedBox(height: 16),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: badges.map((b) {
              final info = _kBadgeLabels[b] ?? (b, Icons.star_rounded);
              return Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: AppColors.primaryLight,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(info.$2, size: 16, color: AppColors.primary),
                    const SizedBox(width: 6),
                    Text(info.$1, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.primary)),
                  ],
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }
}

class _ReferralCard extends ConsumerStatefulWidget {
  final String code;
  final String token;
  const _ReferralCard({required this.code, required this.token});

  @override
  ConsumerState<_ReferralCard> createState() => _ReferralCardState();
}

class _ReferralCardState extends ConsumerState<_ReferralCard> {
  final _controller = TextEditingController();
  bool _redeeming = false;
  String? _msg;

  Future<void> _redeem() async {
    final code = _controller.text.trim();
    if (code.isEmpty) return;
    setState(() { _redeeming = true; _msg = null; });
    final ok = await ApiService.redeemReferral(token: widget.token, code: code);
    if (mounted) setState(() { _redeeming = false; _msg = ok ? 'You both earned 10 credits!' : 'Invalid or already used code.'; });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.border),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 4))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Invite Friends', style: AppTextStyles.cardTitle.copyWith(fontSize: 15)),
          const SizedBox(height: 4),
          Text('Share your code and both earn 10 credits.', style: AppTextStyles.bodySmall),
          const SizedBox(height: 14),
          if (widget.code.isNotEmpty)
            GestureDetector(
              onTap: () {
                Clipboard.setData(ClipboardData(text: widget.code));
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Code copied!')));
              },
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                decoration: BoxDecoration(
                  color: AppColors.surfaceGray,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  children: [
                    Text(widget.code.toUpperCase(), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800, letterSpacing: 3)),
                    const Spacer(),
                    const Icon(Icons.copy_rounded, size: 18, color: AppColors.textSecondary),
                  ],
                ),
              ),
            ),
          const SizedBox(height: 14),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _controller,
                  decoration: InputDecoration(
                    hintText: "Enter friend's code",
                    hintStyle: AppTextStyles.bodySmall,
                    contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.border)),
                    enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.border)),
                    focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: AppColors.primary)),
                  ),
                ),
              ),
              const SizedBox(width: 10),
              ElevatedButton(
                onPressed: _redeeming ? null : _redeem,
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.primary,
                  foregroundColor: Colors.white,
                  elevation: 0,
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                child: _redeeming ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2)) : const Text('Redeem'),
              ),
            ],
          ),
          if (_msg != null) ...[
            const SizedBox(height: 10),
            Text(_msg!, style: TextStyle(fontSize: 13, color: _msg!.contains('earned') ? AppColors.success : AppColors.danger)),
          ],
        ],
      ),
    );
  }
}
