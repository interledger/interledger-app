import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/components.dart';
import 'package:watsonia/routes/personal-details/inquiry_store.dart';
import 'package:watsonia/styles/colors.dart';

class PersonalDetailsRoute extends StatefulWidget {
  const PersonalDetailsRoute({super.key});

  @override
  State<PersonalDetailsRoute> createState() => _PersonalDetailsRouteState();
}

class _PersonalDetailsRouteState extends State<PersonalDetailsRoute> {
  final inquiryStore = InquiryStore()..initializeInquiry();

  @override
  void dispose() {
    super.dispose();
    inquiryStore.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Activate wallet'),
      ),
      body: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          width: double.infinity,
          color: TWColors.bgApp,
          child: SafeArea(
            child: Column(
              children: [
                CustomCard(
                  children: [
                    const CardContent(
                        child: Text(
                            'Complete theses steps to confirm your identity and activate your wallet:')),
                    Padding(
                      padding: const EdgeInsets.fromLTRB(0, 24, 0, 0),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Container(
                            width: 32,
                            height: 32,
                            decoration: const BoxDecoration(
                                color: TWColors.slate300,
                                borderRadius: BorderRadius.all(
                                    Radius.circular(32))),
                          ),
                          Container(
                            margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                            width: 32,
                            height: 32,
                            decoration: const BoxDecoration(
                              color: TWColors.indigo400,
                              borderRadius: BorderRadius.only(
                                bottomLeft: Radius.circular(32),
                                topRight: Radius.circular(32),
                              ),
                            ),
                          ),
                          Flexible(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Personal details',
                                  style: GoogleFonts.inter(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  'Confirmation of your personal details and your address.',
                                  style: GoogleFonts.inter(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 12,
                                        fontWeight: FontWeight.w400),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                Observer(
                  builder: (_) => FilledButton(
                      onPressed: inquiryStore.isInitializing
                          ? null
                          : inquiryStore.startInquiry,
                      child: const Text('Continue')),
                ),
              ],
            ),
          )),
    );
  }
}
