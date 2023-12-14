import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/routes/settings.profile-public.name/form_store.dart';
import 'package:watsonia/stores/wallet_store.dart';
import 'package:watsonia/styles/colors.dart';

class SettingsProfilePublicNameRoute extends StatefulWidget {
  const SettingsProfilePublicNameRoute({super.key});

  @override
  State<SettingsProfilePublicNameRoute> createState() =>
      _SettingsProfilePublicNameRouteState();
}

class _SettingsProfilePublicNameRouteState
    extends State<SettingsProfilePublicNameRoute> {
  final FormStore store = FormStore();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Edit public name'),
      ),
      body: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          width: double.infinity,
          color: TWColors.bgApp,
          child: SafeArea(
            child: Form(
              child: Column(
                children: [
                  AutofillGroup(
                    child: CustomCard(
                      children: [
                        Consumer<WalletStore>(
                          builder: (_, wallet, __) => Observer(
                            builder: (_) => FynbosTextField(
                              onChanged: (value) => store.name = value,
                              labelText: 'Public name',
                              initialValue: wallet.publicName,
                              errorText: store.error.name,
                              textInputAction: TextInputAction.next,
                              keyboardType: TextInputType.text,
                              autofillHints: const [AutofillHints.username],
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  FilledButton(
                      onPressed: () {
                        store.submit(context.read<WalletStore>());
                      },
                      child: const Text('Save')),
                ],
              ),
            ),
          )),
    );
  }
}
