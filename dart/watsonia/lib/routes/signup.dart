import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../styles/colors.dart';

class SignupPage extends StatelessWidget {
  const SignupPage({
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
        width: 100,
        color: TWColors.bgApp,
        child: SafeArea(
          child: Column(
            children: [
              Container(
                  height: 64,
                  width: double.infinity,
                  alignment: Alignment.centerLeft,
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.center,
                    children: const [
                      Padding(
                        padding: EdgeInsets.all(16.0),
                        child: Icon(Icons.menu),
                      ),
                      Image(
                        height: 32,
                        image: NetworkImage(
                            'https://cdn.fynbos.app/logos/32px.png'),
                      ),
                    ],
                  )),
              Container(
                decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(20),
                    color: Colors.white),
                width: double.infinity,
                margin: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      "Sign up",
                      style: GoogleFonts.poppins(
                        textStyle: const TextStyle(
                            color: Color(0xFF0F172A),
                            fontSize: 24,
                            fontWeight: FontWeight.w500),
                      ),
                    ),
                    Text(
                      "Here's what we will need to create your account.",
                      style: GoogleFonts.poppins(
                        textStyle: const TextStyle(
                            decoration: TextDecoration.none,
                            color: Color(0xFF0F172A),
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
                                  "Personal details",
                                  style: GoogleFonts.poppins(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  "Your legal name, email address and country of residence.",
                                  style: GoogleFonts.poppins(
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
                                  "Mobile phone number",
                                  style: GoogleFonts.poppins(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  "A mobile phone number we can verify.",
                                  style: GoogleFonts.poppins(
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
                                  "Password",
                                  style: GoogleFonts.poppins(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  "Create a password we can verify.",
                                  style: GoogleFonts.poppins(
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
                                  "Payment pointer",
                                  style: GoogleFonts.poppins(
                                    textStyle: const TextStyle(
                                        color: Color(0xFF0F172A),
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500),
                                  ),
                                ),
                                Text(
                                  "A mobile phone number we can verify.",
                                  style: GoogleFonts.poppins(
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
                          onPressed: () => {},
                          child: const Text("Let's get started")),
                    )
                  ],
                ),
              ),
            ],
          ),
        ));
  }
}
