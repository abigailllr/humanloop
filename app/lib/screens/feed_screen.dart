import 'package:flutter/material.dart';
import '../models/challenge.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';
import '../widgets/challenge_card.dart';
import 'camera_screen.dart';

final _demo = [
  Challenge(id: 'c1', title: 'Pick & Place', description: 'Pick up any object and place it into a box.', submissions: 1243, difficulty: 'Easy'),
  Challenge(id: 'c2', title: 'Fold It', description: 'Fold a piece of cloth or paper in half.', submissions: 876, difficulty: 'Medium'),
  Challenge(id: 'c3', title: 'Sort & Stack', description: 'Sort 5 objects by size from smallest to largest.', submissions: 542, difficulty: 'Hard'),
  Challenge(id: 'c4', title: 'Pour & Fill', description: 'Pour water from one container to another without spilling.', submissions: 2101, difficulty: 'Easy'),
  Challenge(id: 'c5', title: 'Open & Close', description: 'Open a jar, take something out, and close it again.', submissions: 389, difficulty: 'Medium'),
];

class FeedScreen extends StatelessWidget {
  const FeedScreen({super.key});

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
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text('HumanLoop', style: AppTextStyles.screenTitle),
                      Container(
                        width: 36,
                        height: 36,
                        decoration: const BoxDecoration(color: AppColors.primary, shape: BoxShape.circle),
                        child: const Icon(Icons.person, color: Colors.white, size: 20),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text('Film challenges. Train robots.', style: AppTextStyles.bodyMedium),
                  const SizedBox(height: 24),
                  const _StatsRow(),
                  const SizedBox(height: 28),
                  Text('Active Challenges', style: AppTextStyles.sectionTitle),
                  const SizedBox(height: 4),
                  Text('Film yourself. Earn credits.', style: AppTextStyles.caption),
                ],
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            sliver: SliverList(
              delegate: SliverChildBuilderDelegate(
                (ctx, i) => ChallengeCard(
                  challenge: _demo[i],
                  onTap: () => Navigator.push(ctx, MaterialPageRoute(builder: (_) => CameraScreen(challenge: _demo[i]))),
                ),
                childCount: _demo.length,
              ),
            ),
          ),
          const SliverPadding(padding: EdgeInsets.only(bottom: 16)),
        ],
      ),
    );
  }
}

class _StatsRow extends StatelessWidget {
  const _StatsRow();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: const [
        _StatCard(value: '5,151', label: 'Submissions'),
        SizedBox(width: 12),
        _StatCard(value: '4', label: 'Your credits'),
        SizedBox(width: 12),
        _StatCard(value: '5', label: 'Challenges'),
      ],
    );
  }
}

class _StatCard extends StatelessWidget {
  final String value;
  final String label;
  const _StatCard({required this.value, required this.label});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 16),
        decoration: BoxDecoration(
          color: AppColors.surfaceGray,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(value, style: AppTextStyles.statValue),
            const SizedBox(height: 2),
            Text(label, style: AppTextStyles.caption),
          ],
        ),
      ),
    );
  }
}
