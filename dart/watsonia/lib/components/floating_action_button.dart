import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:grpc/grpc.dart';
import 'package:watsonia/main.dart';
import 'package:watsonia/styles/colors.dart';

import '../generated/protobuf-dart/backend/v1/backend.pbgrpc.dart';
import '../services/grpc_client.dart';

class PayFAB extends StatelessWidget {
  const PayFAB({super.key});

  @override
  Widget build(BuildContext context) {
    return FloatingActionButton.large(
      heroTag: 'main_fab',
      backgroundColor: TWColors.blue[500],
      foregroundColor: TWColors.white,
      // Push goes nested, go is for replacing root pages
      onPressed: () async {
        BackendServiceClient backend = GrpcClient.backend;
        final response = await backend.getPublicWalletInfo(
          GetPublicWalletInfoRequest(
              walletAddress: 'https://local.fynbos.me/Roberto'),
          options: CallOptions(),
        );

        final response3 = await backend.getWalletInfo(
          Empty(),
          options: CallOptions(),
        );
        print('formattedURL from getWalletInfo: ${response3.formattedURL}');
        // await 5s timeout

        print('wallet ID: ${response.walletID}');
        GrpcClient.updateAuthToken('The new token');

        BackendServiceClient backend2 = GrpcClient.backend;
        final response2 = await backend2.getPublicWalletInfo(
          GetPublicWalletInfoRequest(
              walletAddress: 'https://local.fynbos.me/Roberto'),
          options: CallOptions(),
        );
        print('wallet ID 2: ${response2.walletID}');
        const snackAnimation = Animation;
        const snackBar = SnackBar(
          showCloseIcon: true,
          behavior: SnackBarBehavior.floating,
          margin: EdgeInsets.all(16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.all(Radius.circular(12)),
          ),
          content: Text('Yay! A SnackBar!'),
        );

// Find the ScaffoldMessenger in the widget tree
// and use it to show a SnackBar.
        snackbarKey.currentState?.showSnackBar(snackBar);
        // ScaffoldMessenger.of(context).showSnackBar(snackBar);
//         Provider.of<Auth>(context, listen: false).login("username", "password");
        // context.read<Auth>().whoami();
        context.go('/');
      },
      // onPressed: () => context.push('/pay'),
      tooltip: 'Pay',
      child: const Icon(Icons.attach_money_outlined),
    );
  }
}

// TODO This should just trigger the showDialog in the dialog.dart example
