import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../theme/colors.dart';
import '../theme/text_styles.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  int _credits = 0;
  int _submissions = 0;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final profile = await ApiService.getProfile();
      setState(() {
        _credits = profile['credits'] ?? _credits;
        _submissions = profile['submissions'] ?? _submissions;
      });
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(24, 24, 24, 32),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Profile', style: AppTextStyles.screenTitle),
            const SizedBox(height: 28),
            _AvatarCard(credits: _credits, submissions: _submissions),
            const SizedBox(height: 20),
            _BalanceCard(credits: _credits),
            const SizedBox(height: 20),
            _ActivityCard(submissions: _submissions),
            const SizedBox(height: 28),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () {},
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
          ],
        ),
      ),
    );
  }
}

class _AvatarCard extends StatelessWidget {
  final int credits;
  final int submissions;
  const _AvatarCard({required this.credits, required this.submissions});

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
                Text('Contributor', style: AppTextStyles.cardTitle),
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
  const _ActivityCard({required this.submissions});

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
          _ActivityRow(icon: Icons.check_circle_outline, label: 'Verified', value: '${(submissions * 0.8).round()}'),
          const SizedBox(height: 14),
          _ActivityRow(icon: Icons.bolt_outlined, label: 'Challenges Done', value: '5'),
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
