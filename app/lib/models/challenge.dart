class Challenge {
  final String id;
  final String title;
  final String description;
  final int submissions;
  final String difficulty;

  const Challenge({
    required this.id,
    required this.title,
    required this.description,
    required this.submissions,
    this.difficulty = 'Easy',
  });

  factory Challenge.fromJson(Map<String, dynamic> j) => Challenge(
        id: j['id'],
        title: j['title'],
        description: j['description'],
        submissions: j['submissions'] ?? 0,
        difficulty: j['difficulty'] ?? 'Easy',
      );
}
