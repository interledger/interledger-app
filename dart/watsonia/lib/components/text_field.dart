import 'package:flutter/material.dart';
import 'package:watsonia/styles/colors.dart';

class FynbosTextField extends StatelessWidget {
  final String labelText;
  final String? initialValue;
  final String? errorText;
  final bool obscureText;
  final TextEditingController? controller;
  final void Function(String)? onChanged;
  final TextInputType keyboardType;
  final TextCapitalization textCapitalization;
  final TextStyle textStyle;
  final TextInputAction? textInputAction;
  final Iterable<String>? autofillHints;

  const FynbosTextField(
      {super.key,
      this.controller,
      required this.labelText,
      this.initialValue,
      this.errorText,
      this.obscureText = false,
      required this.onChanged,
      this.keyboardType = TextInputType.text,
      this.textCapitalization = TextCapitalization.none,
      this.textStyle = const TextStyle(fontSize: 16),
      this.textInputAction,
      this.autofillHints});

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: labelText,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.only(left: 8.0),
            child: Text(
              labelText,
              // TODO specify fixed text styles
              style: const TextStyle(
                color: TWColors.textMedium,
                fontSize: 14,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          const SizedBox(height: 4),
          TextFormField(
            initialValue: initialValue,
            autofillHints: autofillHints,
            textInputAction: textInputAction,
            onChanged: onChanged,
            controller: controller,
            obscureText: obscureText,
            decoration: InputDecoration(
              // suffixIcon: const Icon(Icons.hide_source),
              errorText: errorText,
            ),
          ),
        ],
      ),
    );
  }
}
