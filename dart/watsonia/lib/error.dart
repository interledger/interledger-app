import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/styles/colors.dart';

class ErrorRoute extends StatefulWidget {
  const ErrorRoute({super.key});

  @override
  State<ErrorRoute> createState() => _ErrorRouteState();
}

class _ErrorRouteState extends State<ErrorRoute> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            CustomCard(children: [
              const CardHeader(
                  child: CardTitle(title: 'An error done happened')),
              CardContent(
                  child: Column(
                children: <Widget>[
                  Text(
                    'Share your wallet address to get paid, or click the pay button to transact.',
                    style: GoogleFonts.inter(
                      textStyle: const TextStyle(
                        fontWeight: FontWeight.w400,
                        fontSize: 16,
                      ),
                    ),
                  )
                ],
              )),
            ]),
            FilledButton(
                style: ButtonStyle(
                  minimumSize: MaterialStateProperty.all(
                      const Size(double.infinity, 48)),
                  backgroundColor: MaterialStateProperty.resolveWith<Color?>(
                    (Set<MaterialState> states) {
                      if (states.contains(MaterialState.pressed)) {
                        return TWColors.bgContainerPrimaryActive;
                      }
                      return TWColors.bgPrimary; // Use the component's default.
                    },
                  ),
                ),
                onPressed: () async {
                  context.go('/');
                },
                child: const Text('Go home')),
          ],
        ),
      ),
    );
  }
}
