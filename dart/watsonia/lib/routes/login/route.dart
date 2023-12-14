import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/routes/login/form_store.dart';
import 'package:watsonia/stores/auth_store.dart';
import 'package:watsonia/styles/colors.dart';

class LoginRoute extends StatefulWidget {
  const LoginRoute({super.key});

  @override
  State<LoginRoute> createState() => _LoginRouteState();
}

class _LoginRouteState extends State<LoginRoute> {
  final FormStore store = FormStore();

  @override
  void initState() {
    super.initState();
  }

  @override
  void dispose() {
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Log in'),
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
                        Observer(
                          builder: (_) => FynbosTextField(
                            onChanged: (value) => store.email = value,
                            labelText: 'Email',
                            errorText: store.error.email,
                            textInputAction: TextInputAction.next,
                            keyboardType: TextInputType.text,
                            autofillHints: const [AutofillHints.username],
                          ),
                        ),
                        const SizedBox(height: 8),
                        Observer(
                          builder: (_) => FynbosTextField(
                            onChanged: (value) => store.password = value,
                            labelText: 'Password',
                            obscureText: true,
                            errorText: store.error.password,
                            textInputAction: TextInputAction.done,
                            keyboardType: TextInputType.text,
                            autofillHints: const [AutofillHints.password],
                          ),
                        ),
                      ],
                    ),
                  ),
                  FilledButton(
                      onPressed: () async {
                        store.submit(context.read<AuthStore>());
                      },
                      child: const Text('Log in')),
                ],
              ),
            ),
          )),
    );
  }
}
