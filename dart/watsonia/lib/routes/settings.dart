import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/stores/kyc_store.dart';
import 'package:watsonia/styles/colors.dart';

class SettingsRoute extends StatefulWidget {
  const SettingsRoute({super.key});

  @override
  State<SettingsRoute> createState() => _SettingsRouteState();
}

class _SettingsRouteState extends State<SettingsRoute> {
  @override
  Widget build(BuildContext context) {
    final kycStore = Provider.of<KYCStore>(context)..init();
    return Scaffold(
      // We don't need an app bar on this page because the shell route scaffold will render it for us.
      appBar: AppBar(
        title: const Text('Settings'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          CustomCard(
            children: <Widget>[
              const CardHeader(child: CardTitle(title: 'Profile')),
              if (kycStore.status != 0)
                CardLink(
                  onPressed: () {
                    context.go('/settings/profile-personal');
                  },
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.account_circle_outlined),
                          const SizedBox(width: 12),
                          Text(
                            'Personal information',
                            style: GoogleFonts.inter(
                              textStyle: const TextStyle(
                                fontWeight: FontWeight.w400,
                                fontSize: 16,
                              ),
                            ),
                          ),
                        ],
                      ),
                      const Icon(Icons.navigate_next)
                    ],
                  ),
                ),
              CardLink(
                onPressed: () {
                  context.push('/settings/profile-public');
                },
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.contact_page_outlined),
                        const SizedBox(width: 12),
                        Text(
                          'Public information',
                          style: GoogleFonts.inter(
                            textStyle: const TextStyle(
                              fontWeight: FontWeight.w400,
                              fontSize: 16,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const Icon(Icons.navigate_next)
                  ],
                ),
              ),
              CardLink(
                onPressed: () {
                  context.go('/settings/profile-contact');
                },
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.call_outlined),
                        const SizedBox(width: 12),
                        Text(
                          'Contact information',
                          style: GoogleFonts.inter(
                            textStyle: const TextStyle(
                              fontWeight: FontWeight.w400,
                              fontSize: 16,
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
          CustomCard(
            children: <Widget>[
              const CardHeader(child: CardTitle(title: 'Security')),
              CardLink(
                onPressed: () {
                  context.go('/logout');
                },
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.logout),
                        const SizedBox(width: 12),
                        Text(
                          'Logout',
                          style: GoogleFonts.inter(
                            textStyle: const TextStyle(
                              fontWeight: FontWeight.w400,
                              fontSize: 16,
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
          )
        ],
      ),
      drawerScrimColor: TWColors.bgScrim,
      drawer: const NavDrawer(),
      floatingActionButton: const PayFAB(),
    );
  }
}
