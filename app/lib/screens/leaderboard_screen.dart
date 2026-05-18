import 'package:flutter/material.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../models/user.dart';
import '../widgets/leaderboard_tile.dart';

final _demoLeaders = [
  AppUser(id: 'u1', name: 'Maria Santos',   email: '', credits: 312, submissions: 87, rank: 1),
  AppUser(id: 'u2', name: 'James Kim',      email: '', credits: 278, submissions: 74, rank: 2),
  AppUser(id: 'u3', name: 'Priya Nair',     email: '', credits: 241, submissions: 69, rank: 3),
  AppUser(id: 'u4', name: 'Luca Ferrari',   email: '', credits: 198, submissions: 55, rank: 4),
  AppUser(id: 'u5', name: 'Aiko Tanaka',    email: '', credits: 175, submissions: 49, rank: 5),
  AppUser(id: 'u6', name: 'Omar Hassan',    email: '', credits: 154, submissions: 44, rank: 6),
  AppUser(id: 'u7', name: 'Sophie Dubois',  email: '', credits: 132, submissions: 38, rank: 7),
  AppUser(id: 'u8', name: 'Raj Patel',      email: '', credits: 110, submissions: 32, rank: 8),
  AppUser(id: 'u9', name: 'Elena Kovacs',   email: '', credits: 89,  submissions: 25, rank: 9),
  AppUser(id: 'u10', name: 'Alex Johnson',  email: '', credits: 4,   submissions: 3,  rank: 42),
];

class LeaderboardScreen extends StatelessWidget {
  const LeaderboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final top3 = _demoLeaders.take(3).toList();
    final rest = _demoLeaders.skip(3).toList();

    return SafeArea(
      child: CustomScrollView(
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
                  _Podium(top3: top3),
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
      ),
    );
  }
}

class _Podium extends StatelessWidget {
  final List<AppUser> top3;
  const _Podium({required this.top3});

  @override
  Widget build(BuildContext context) {
    // Order: 2nd, 1st, 3rd
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
