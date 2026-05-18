import 'package:flutter/material.dart';
import '../models/user.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';

class LeaderboardTile extends StatelessWidget {
  final AppUser user;
  final bool isCurrentUser;

  const LeaderboardTile({super.key, required this.user, this.isCurrentUser = false});

  @override
  Widget build(BuildContext context) {
    final highlight = isCurrentUser || user.name == 'Alex Johnson';

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: highlight ? AppColors.primaryLight : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: highlight ? AppColors.primary.withValues(alpha: 0.2) : AppColors.border),
        boxShadow: highlight
            ? [BoxShadow(color: AppColors.primary.withValues(alpha: 0.08), blurRadius: 12, offset: const Offset(0, 4))]
            : [BoxShadow(color: Colors.black.withValues(alpha: 0.03), blurRadius: 8, offset: const Offset(0, 2))],
      ),
      child: Row(
        children: [
          // Rank
          SizedBox(
            width: 32,
            child: Text(
              '#${user.rank}',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w700,
                color: highlight ? AppColors.primary : AppColors.textSecondary,
              ),
            ),
          ),
          const SizedBox(width: 12),
          // Avatar
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: highlight ? AppColors.primary : AppColors.surfaceGray,
              shape: BoxShape.circle,
            ),
            child: Center(
              child: Text(
                user.name.substring(0, 1),
                style: TextStyle(
                  fontSize: 17,
                  fontWeight: FontWeight.w700,
                  color: highlight ? Colors.white : AppColors.textSecondary,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          // Name + submissions
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(user.name, style: AppTextStyles.cardTitle.copyWith(fontSize: 14)),
                    if (highlight) ...[
                      const SizedBox(width: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(color: AppColors.primary, borderRadius: BorderRadius.circular(4)),
                        child: const Text('You', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: Colors.white)),
                      ),
                    ],
                  ],
                ),
                const SizedBox(height: 2),
                Text('${user.submissions} submissions', style: AppTextStyles.caption),
              ],
            ),
          ),
          // Credits
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text('${user.credits}', style: AppTextStyles.statValue.copyWith(fontSize: 18)),
              Text('credits', style: AppTextStyles.caption),
            ],
          ),
        ],
      ),
    );
  }
}
