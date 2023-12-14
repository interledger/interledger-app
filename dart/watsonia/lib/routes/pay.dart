import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/components.dart';

class PayRoute extends StatefulWidget {
  const PayRoute({super.key});

  @override
  State<PayRoute> createState() => _PayPagePageState();
}

class _PayPagePageState extends State<PayRoute> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Image.asset(
          'images/Logo.png',
          height: 32,
        ),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          CustomCard(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
                child: Text(
                  'Pay',
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
    );
  }
}
