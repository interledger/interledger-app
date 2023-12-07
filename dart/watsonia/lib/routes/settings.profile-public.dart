import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/stores/wallet_store.dart';

class SettingsProfilePublicRoute extends StatefulWidget {
  const SettingsProfilePublicRoute({super.key});

  @override
  State<SettingsProfilePublicRoute> createState() =>
      _SettingsProfilePublicRouteState();
}

class _SettingsProfilePublicRouteState
    extends State<SettingsProfilePublicRoute> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      // We don't need an app bar on this page because the shell route scaffold will render it for us.
      appBar: AppBar(
        title: const Text('Public information'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          const CustomCard(
            children: <Widget>[
              CardContent(
                  child: Text(
                      'The following information will appear on your public Fynbos.me page.')),
            ],
          ),
          CustomCard(
            children: <Widget>[
              CardLink(
                onPressed: () {
                  context.go('/settings/profile-public/name');
                },
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.account_circle_outlined),
                        const SizedBox(width: 12),
                        Consumer<WalletStore>(
                          builder: (_, wallet, __) => Observer(
                            builder: (_) => Text(
                              wallet.publicName,
                              style: GoogleFonts.inter(
                                textStyle: const TextStyle(
                                  fontWeight: FontWeight.w400,
                                  fontSize: 16,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const Icon(Icons.navigate_next)
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
