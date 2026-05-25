class AppUser {
  final String id;
  final String name;
  final String email;
  final String? avatarUrl;
  final int credits;
  final int submissions;
  final int rank;
  final String level;
  final List<String> badges;
  final String referralCode;

  const AppUser({
    required this.id,
    required this.name,
    required this.email,
    this.avatarUrl,
    this.credits = 0,
    this.submissions = 0,
    this.rank = 0,
    this.level = 'Rookie',
    this.badges = const [],
    this.referralCode = '',
  });

  AppUser copyWith({
    String? name,
    String? email,
    String? avatarUrl,
    int? credits,
    int? submissions,
    int? rank,
    String? level,
    List<String>? badges,
    String? referralCode,
  }) {
    return AppUser(
      id: id,
      name: name ?? this.name,
      email: email ?? this.email,
      avatarUrl: avatarUrl ?? this.avatarUrl,
      credits: credits ?? this.credits,
      submissions: submissions ?? this.submissions,
      rank: rank ?? this.rank,
      level: level ?? this.level,
      badges: badges ?? this.badges,
      referralCode: referralCode ?? this.referralCode,
    );
  }

  factory AppUser.fromJson(Map<String, dynamic> j) => AppUser(
        id: j['id'] as String,
        name: j['name'] as String,
        email: j['email'] as String,
        avatarUrl: j['avatar_url'] as String?,
        credits: j['credits'] as int? ?? 0,
        submissions: j['submissions'] as int? ?? 0,
        rank: j['rank'] as int? ?? 0,
        level: j['level'] as String? ?? 'Rookie',
        badges: (j['badges'] as List<dynamic>?)?.cast<String>() ?? [],
        referralCode: j['referral_code'] as String? ?? '',
      );
}
