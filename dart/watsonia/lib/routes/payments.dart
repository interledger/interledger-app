import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/styles/colors.dart';

class PaymentsRoute extends StatefulWidget {
  const PaymentsRoute({super.key});

  @override
  State<PaymentsRoute> createState() => _PaymentsRouteState();
}

class _PaymentsRouteState extends State<PaymentsRoute> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Payments'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          CustomCard(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
                child: Text(
                  'Payments',
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
      floatingActionButton: const PayFAB(),
    );
  }
}
