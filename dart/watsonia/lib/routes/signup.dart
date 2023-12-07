import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/components/card.dart' as comp;

import '../styles/colors.dart';

class SignupRoute extends StatelessWidget {
  const SignupRoute({super.key});

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
                comp.CustomCard(
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
                    Padding(
                      padding: const EdgeInsets.fromLTRB(0, 24, 0, 0),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Container(
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                                color: TWColors.yellow[300],
                                borderRadius: const BorderRadius.only(
                                    topLeft: Radius.circular(32))),
                          ),
                          Container(
                            margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                              color: TWColors.slate[500],
                              borderRadius: const BorderRadius.only(
                                bottomLeft: Radius.circular(32),
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
                                  'Your legal name, email address and country of residence.',
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
                    Padding(
                      padding: const EdgeInsets.fromLTRB(0, 32, 0, 0),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Container(
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                                color: TWColors.rose[400],
                                borderRadius: const BorderRadius.only(
                                    bottomLeft: Radius.circular(32))),
                          ),
                          Container(
                            margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                              color: TWColors.lime[500],
                              borderRadius: BorderRadius.circular(32),
                            ),
                          ),
                          Flexible(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Mobile phone number',
                                  style: GoogleFonts.inter(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  'A mobile phone number we can verify.',
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
                    Padding(
                      padding: const EdgeInsets.fromLTRB(0, 32, 0, 0),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Container(
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                                color: TWColors.yellow[300],
                                borderRadius: const BorderRadius.only(
                                    topLeft: Radius.circular(32))),
                          ),
                          Container(
                            margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                              color: TWColors.slate[300],
                              borderRadius: const BorderRadius.only(
                                  bottomLeft: Radius.circular(32)),
                            ),
                          ),
                          Flexible(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Password',
                                  style: GoogleFonts.inter(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  'Create a password we can verify.',
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
                    Padding(
                      padding: const EdgeInsets.fromLTRB(0, 32, 0, 0),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Container(
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                                color: TWColors.rose[500],
                                borderRadius: BorderRadius.circular(32)),
                          ),
                          Container(
                            margin: const EdgeInsets.fromLTRB(0, 0, 16, 0),
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                              color: TWColors.lime[300],
                              borderRadius: const BorderRadius.only(
                                  topRight: Radius.circular(32)),
                            ),
                          ),
                          Flexible(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Payment pointer',
                                  style: GoogleFonts.inter(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  'A mobile phone number we can verify.',
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
                    Container(
                      margin: const EdgeInsets.fromLTRB(0, 32, 0, 0),
                      width: double.infinity,
                      child: FilledButton(
                          style: ButtonStyle(
                            backgroundColor:
                                MaterialStateProperty.resolveWith<Color?>(
                              (Set<MaterialState> states) {
                                if (states.contains(MaterialState.pressed)) {
                                  return Theme.of(context)
                                      .colorScheme
                                      .primary
                                      .withOpacity(0.5);
                                }
                                return TWColors
                                    .bgPrimary; // Use the component's default.
                              },
                            ),
                          ),
                          onPressed: () async {
                            // Provider.of<Auth>(context, listen: false)
                            //     .login('username', 'password');
                            // context.read<Auth>().whoami();
                            context.go('/');
                          },
                          child: const Text("Let's get started")),
                    )
                  ],
                ),
              ],
            ),
          )),
    );
  }
}
