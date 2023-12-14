import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/styles/colors.dart';

class SupportRoute extends StatefulWidget {
  const SupportRoute({super.key});

  @override
  State<SupportRoute> createState() => _SupportRouteState();
}

class _SupportRouteState extends State<SupportRoute> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Support'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          CustomCard(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
                child: Text(
                  'Transactions',
                  style: GoogleFonts.inter(
                    textStyle: const TextStyle(
                      fontWeight: FontWeight.w500,
                      fontSize: 24,
                    ),
                  ),
                ),
              ),
            ],
          )
        ],
      ),
      drawerScrimColor: TWColors.bgScrim,
      drawer: const NavDrawer(),
    );
  }
}
