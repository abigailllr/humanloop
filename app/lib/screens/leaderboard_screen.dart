import 'package:flutter/material.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../models/user.dart';
import '../services/api_service.dart';
import '../widgets/leaderboard_tile.dart';

class LeaderboardScreen extends StatefulWidget {
  const LeaderboardScreen({super.key});

  @override
  State<LeaderboardScreen> createState() => _LeaderboardScreenState();
}

class _LeaderboardScreenState extends State<LeaderboardScreen> {
  late Future<List<AppUser>> _leaders;

  @override
  void initState() {
    super.initState();
    _leaders = ApiService.getLeaderboard();
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: FutureBuilder<List<AppUser>>(
        future: _leaders,
        builder: (ctx, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }
          final all = snap.data ?? [];
          final top3 = all.take(3).toList();
          final rest = all.skip(3).toList();

          return CustomScrollView(
            slivers: [
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(24, 24, 24, 8),
                sliver: SliverToBoxAdapter(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Leaderboard', style: AppTextStyles.screenTitle),
                      const SizedBox(height: 4),
                      Text('Top contributors this month', style: AppTextStyles.bodySmall),
                      const SizedBox(height: 28),
                      if (top3.length == 3) _Podium(top3: top3),
                      const SizedBox(height: 28),
                      Text('Rankings', style: AppTextStyles.sectionTitle),
                      const SizedBox(height: 4),
                    ],
                  ),
                ),
              ),
              SliverPadding(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate(
                    (_, i) => LeaderboardTile(user: rest[i]),
                    childCount: rest.length,
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

class _Podium extends StatelessWidget {
  final List<AppUser> top3;
  const _Podium({required this.top3});

  @override
  Widget build(BuildContext context) {
    final order = [top3[1], top3[0], top3[2]];
    final heights = [80.0, 110.0, 60.0];
    final ranks = [2, 1, 3];
    final medals = ['🥈', '🥇', '🥉'];

    return Row(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: List.generate(3, (i) {
        final user = order[i];
        return Expanded(
          child: Column(
            children: [
              Text(medals[i], style: const TextStyle(fontSize: 28)),
              const SizedBox(height: 6),
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: AppColors.primaryLight,
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: ranks[i] == 1 ? const Color(0xFFFBBC05) : AppColors.border,
                    width: ranks[i] == 1 ? 2 : 1,
                  ),
                ),
                child: Center(
                  child: Text(
                    user.name.substring(0, 1),
                    style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: AppColors.primary),
                  ),
                ),
              ),
              const SizedBox(height: 6),
              Text(user.name.split(' ').first, style: AppTextStyles.label.copyWith(color: AppColors.textPrimary, fontWeight: FontWeight.w600, fontSize: 12), textAlign: TextAlign.center),
              Text('${user.credits} cr', style: AppTextStyles.caption),
              const SizedBox(height: 8),
              Container(
                height: heights[i],
                decoration: BoxDecoration(
                  color: ranks[i] == 1 ? AppColors.primary : AppColors.surfaceGray,
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
                ),
                child: Center(
                  child: Text(
                    '#${ranks[i]}',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                      color: ranks[i] == 1 ? Colors.white : AppColors.textSecondary,
                    ),
                  ),
                ),
              ),
            ],
          ),
        );
      }),
    );
  }
}
