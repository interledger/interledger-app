import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/stores/kyc_store.dart';
import 'package:watsonia/stores/wallet_store.dart';
import 'package:watsonia/styles/colors.dart';

class HomeRoute extends StatefulWidget {
  const HomeRoute({super.key});

  @override
  State<HomeRoute> createState() => _HomeRouteState();
}

class _HomeRouteState extends State<HomeRoute> {
  @override
  Widget build(BuildContext context) {
    final kycStore = Provider.of<KYCStore>(context)..init();
    return Scaffold(
      appBar: AppBar(
        title: const Text('Home'),
        actions: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                    color: TWColors.yellow[300],
                    borderRadius:
                        const BorderRadius.only(topLeft: Radius.circular(24))),
              ),
              Container(
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                    color: TWColors.rose[300],
                    borderRadius: const BorderRadius.only(
                        bottomLeft: Radius.circular(24))),
              ),
              Container(
                margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                  color: TWColors.slate[500],
                  borderRadius: const BorderRadius.only(
                      bottomLeft: Radius.circular(24),
                      bottomRight: Radius.circular(24)),
                ),
              ),
            ],
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
        children: <Widget>[
          CustomCard(children: [
            const CardHeader(child: CardTitle(title: 'Wallet')),
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
            Consumer<WalletStore>(
              builder: (_, wallet, __) => Observer(
                builder: (_) => Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(16),
                  decoration: const BoxDecoration(
                    color: TWColors.bgContainer,
                    borderRadius: BorderRadius.all(Radius.circular(12)),
                  ),
                  child: Text(
                    // TODO Replace this with the wallet address
                    wallet.walletAddress,
                    style: GoogleFonts.inter(
                      textStyle: const TextStyle(
                        fontWeight: FontWeight.w400,
                        fontSize: 16,
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ]),
        ],
      ),
      drawerScrimColor: TWColors.bgScrim,
      drawer: const NavDrawer(),
      floatingActionButton: const PayFAB(),
    );
  }
}
