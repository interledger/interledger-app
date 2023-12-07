// import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/routes/login/form_store.dart';
import 'package:watsonia/styles/colors.dart';
import 'package:watsonia/components/text_field.dart';

import '../../stores/auth_store.dart';

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
    // TODO Instead of setupValidations we should init a login flow
    // store.setupValidations();
  }

  @override
  void dispose() {
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final authStore = Provider.of<AuthStore>(context);
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
                      style: ButtonStyle(
                        minimumSize: MaterialStateProperty.all(
                            const Size(double.infinity, 48)),
                        backgroundColor:
                            MaterialStateProperty.resolveWith<Color?>(
                          (Set<MaterialState> states) {
                            if (states.contains(MaterialState.pressed)) {
                              return TWColors.bgContainerPrimaryActive;
                            }
                            return TWColors
                                .bgPrimary; // Use the component's default.
                          },
                        ),
                      ),
                      onPressed: () async {
                        // TODO Move the form store to be a child of the auth store.
                        //  This will allow us to call the login function in the auth store directly and pass the BuildContext to the login function
                        //  - to route after successfully logging in
                        //  - to pass errors back to the UI.
                        store.submit(authStore);
                        // Provider.of<Auth>(context, listen: false)
                        //     .login("username", "password");
                        // context.go('/');
                      },
                      child: const Text('Log in')),
                ],
              ),
            ),
          )),
    );
  }
}
