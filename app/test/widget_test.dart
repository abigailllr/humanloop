import 'package:flutter_test/flutter_test.dart';
import 'package:humanloop/main.dart';

void main() {
  testWidgets('app launches', (WidgetTester tester) async {
    await tester.pumpWidget(const HumanLoopApp());
    expect(find.text('HumanLoop'), findsOneWidget);
  });
}
