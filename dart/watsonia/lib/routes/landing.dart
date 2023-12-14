import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/stores/auth_store.dart';

import '../styles/colors.dart';

class LandingRoute extends StatelessWidget {
  const LandingRoute({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Image.asset(
          'images/Logo.png',
          height: 32,
        ),
      ),
      body: Container(
          width: double.infinity,
          color: TWColors.bgApp,
          child: SafeArea(
            child: Column(
              children: [
                CustomCard(
                  children: [
                    Text(
                      "Here's what we will need to create your account:",
                      style: GoogleFonts.inter(
                        textStyle: const TextStyle(
                            color: TWColors.textStrong,
                            fontSize: 16,
                            fontWeight: FontWeight.w400),
                      ),
                    ),
                    FilledButton(
                        onPressed: () async {
                          Provider.of<AuthStore>(context, listen: false)
                              .initLoginFlow();
                          context.push('/login');
                        },
                        child: const Text('Log in')),
                  ],
                ),
              ],
            ),
          )),
    );
  }
}
