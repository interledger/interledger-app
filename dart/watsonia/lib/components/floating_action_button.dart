import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/styles/colors.dart';

class PayFAB extends StatelessWidget {
  const PayFAB({super.key});

  @override
  Widget build(BuildContext context) {
    return FloatingActionButton.large(
      heroTag: 'main_fab',
      backgroundColor: TWColors.bgPrimary,
      foregroundColor: TWColors.white,
      onPressed: () => showDialog<void>(
        barrierDismissible: true,
        barrierColor: TWColors.bgScrim,
        context: context,
        builder: (BuildContext context) {
          return Column(
            mainAxisAlignment: MainAxisAlignment.start,
            children: <Widget>[
              Dialog(
                backgroundColor: TWColors.bgPage,
                shape: const RoundedRectangleBorder(
                  borderRadius: BorderRadius.all(Radius.circular(20)),
                ),
                insetPadding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    Padding(
                      padding: const EdgeInsets.all(8.0),
                      child: Observer(
                        builder: (_) => FynbosTextField(
                          onChanged: (value) => print(value),
                          // errorText: store.error.password,
                          autofocus: true,
                          textInputAction: TextInputAction.search,
                          keyboardType: TextInputType.text,
                          autofillHints: const [AutofillHints.password],
                        ),
                      ),
                    ),
                  ],
                ),
              )
            ],
          );
        },
      ),
      tooltip: 'Pay',
      child: const Icon(Icons.attach_money_outlined),
    );
  }
}
