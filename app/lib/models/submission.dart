class Submission {
  final String id;
  final String challengeId;
  final String challengeTitle;
  final DateTime submittedAt;
  final SubmissionStatus status;
  final int creditsEarned;
  final double qualityScore;

  const Submission({
    required this.id,
    required this.challengeId,
    required this.challengeTitle,
    required this.submittedAt,
    this.status = SubmissionStatus.pending,
    this.creditsEarned = 0,
    this.qualityScore = 0,
  });

  factory Submission.fromJson(Map<String, dynamic> j) => Submission(
        id: j['id'] as String,
        challengeId: j['challenge_id'] as String,
        challengeTitle: j['challenge_title'] as String? ?? '',
        submittedAt: DateTime.parse((j['submitted_at'] ?? j['created_at']) as String),
        status: _parseStatus(j['status'] as String? ?? 'pending'),
        creditsEarned: j['credits_earned'] as int? ?? 0,
        qualityScore: (j['quality_score'] as num?)?.toDouble() ?? 0,
      );

  static SubmissionStatus _parseStatus(String s) {
    if (s == 'done') return SubmissionStatus.verified;
    if (s == 'failed') return SubmissionStatus.rejected;
    if (s == 'synthetic') return SubmissionStatus.synthetic;
    return SubmissionStatus.pending;
  }
}

enum SubmissionStatus { pending, verified, rejected, synthetic }
