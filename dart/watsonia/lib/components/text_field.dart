//
// // Convert this component to dart to be used in a flutter project.
// // Path: dart/watsonia/lib/components/text_field.dart
// // Compare this snippet from dart/watsonia/lib/routes/signup.dart:
// //         const SizedBox(height: 40),
// //         TextField(
// //           label: 'Mobile phone number',
// //           labelSuffix: 'Required',
// //           errorMessage: 'Please enter a valid phone number.',
// //           prefixIcon: Icon(Icons.phone_android, color: Colors.black),
// //         ),
// //         const SizedBox(height: 40),
// //         TextField(
// //           label: 'Password',
// //           labelSuffix: 'Required',
// //           errorMessage: 'Please enter a valid password.',
// //           prefixIcon: Icon(Icons.lock, color: Colors.black),
// //         ),
// //         const SizedBox(height: 40),
// //         TextField(
// //           label: 'Payment pointer',
// //           labelSuffix: 'Required',
// //           errorMessage: 'Please enter a valid payment pointer.',
// //           prefixIcon: Icon(Icons.payment, color: Colors.black),
// //         ),
// //         const SizedBox(height: 40),
// //         Button(
// //           label: "Let's get started",
// //           onPressed: () => Navigator.pushNamed(context, '/signup'),
// //         ),
// //         const SizedBox(height: 40),
// //         Row(
// //           mainAxisAlignment: MainAxisAlignment.center,
// //           children: [
// //             Text(
// //               'Already have an account?',
// //               style: TextStyle(
// //                 fontSize: 14,
// //                 color: Colors.black,
// //               ),
// //             ),
// //             TextButton(
// //               onPressed: () => Navigator.pushNamed(context, '/signup'),
// //               child: Text(
// //                 'Log in',
// //                 style: TextStyle(
// //                   fontSize: 14,
// //                   color: Colors.black,
// //                   fontWeight: FontWeight.bold,
// //                 ),
// //               ),
// //             ),
// //           ],
// //         ),
// //       ],
// //     );
// //   }
// // }
// import 'package:flutter/material.dart';
//
// class TextField extends StatelessWidget {
//   const TextField({
//     Key? key,
//     required this.label,
//     this.labelSuffix,
//     this.errorMessage,
//     this.successMessage,
//     this.prefix,
//     this.prefixIcon,
//     this.appendIcon,
//   }) : super(key: key);
//
//   final String label;
//   final String? labelSuffix;
//   final String? errorMessage;
//   final String? successMessage;
//   final String? prefix;
//   final Icon? prefixIcon;
//   final Icon? appendIcon;
//
//   @override
//   Widget build(BuildContext context) {
//     return Container(
//       child: Column(
//         crossAxisAlignment: CrossAxisAlignment.start,
//         children: [
//           Text('$label ${labelSuffix != null ? '<sup>$labelSuffix</sup>' : ''}'),
//           SizedBox(height: 8),
//           Container(
//             decoration: BoxDecoration(
//               borderRadius: BorderRadius.circular(10),
//               border: Border.all(color: Colors.black),
//             ),
//             padding: EdgeInsets.only(left: 16, right: 16),
//             child: Row(
//               children: [
//                 if (prefixIcon != null) prefixIcon!,
//                 if (prefix != null) Text(prefix!),
//                 Expanded(
//                   child: TextField(
//                     label: '',
//
//                     decoration: InputDecoration(
//                       border: InputBorder.none,
//                       hintText: 'Enter your text',
//                     ),
//                   ),
//                 ),
//                 if (appendIcon != null) appendIcon!,
//               ],
//             ),
//           ),
//           SizedBox(height: 8),
//           if (errorMessage != null) Text(errorMessage!),
//           if (successMessage != null && errorMessage == null)
//             Text(successMessage!),
//         ],
//       ),
//     );
//   }
