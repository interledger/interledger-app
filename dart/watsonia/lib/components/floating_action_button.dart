import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:watsonia/styles/colors.dart';

class PayFAB extends StatelessWidget {
  const PayFAB({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return FloatingActionButton.large(
      heroTag: "main_fab",
      backgroundColor: TWColors.blue[500],
      foregroundColor: TWColors.white,
      // Push goes nested, go is for replacing root pages
      onPressed: () => context.push('/pay'),
      tooltip: 'Pay',
      child: const Icon(Icons.attach_money_outlined),
    );
  }
}
