class Submission {
  final String id;
  final String challengeId;
  final String challengeTitle;
  final DateTime submittedAt;
  final SubmissionStatus status;
  final int creditsEarned;

  const Submission({
    required this.id,
    required this.challengeId,
    required this.challengeTitle,
    required this.submittedAt,
    this.status = SubmissionStatus.pending,
    this.creditsEarned = 0,
  });

  factory Submission.fromJson(Map<String, dynamic> j) => Submission(
        id: j['id'] as String,
        challengeId: j['challenge_id'] as String,
        challengeTitle: j['challenge_title'] as String,
        submittedAt: DateTime.parse(j['submitted_at'] as String),
        status: SubmissionStatus.values.firstWhere(
          (s) => s.name == (j['status'] as String? ?? 'pending'),
          orElse: () => SubmissionStatus.pending,
        ),
        creditsEarned: j['credits_earned'] as int? ?? 0,
      );
}

enum SubmissionStatus { pending, verified, rejected }
