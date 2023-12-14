import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:watsonia/styles/colors.dart';

class FynbosTextField extends StatelessWidget {
  final String? labelText;
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
  final bool autofocus;

  const FynbosTextField({
    super.key,
    this.controller,
    this.labelText,
    this.initialValue,
    this.errorText,
    this.obscureText = false,
    required this.onChanged,
    this.keyboardType = TextInputType.text,
    this.textCapitalization = TextCapitalization.none,
    this.textStyle = const TextStyle(fontSize: 16),
    this.textInputAction,
    this.autofillHints,
    this.autofocus = false,
  });

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: labelText,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (labelText != null) ...[
            Padding(
              padding: const EdgeInsets.only(left: 8.0),
              child: Text(
                labelText!,
                // TODO specify fixed text styles
                style: const TextStyle(
                  color: TWColors.textMedium,
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
            const SizedBox(height: 4),
          ],
          TextFormField(
            initialValue: initialValue,
            autofillHints: autofillHints,
            textInputAction: textInputAction,
            onChanged: onChanged,
            controller: controller,
            obscureText: obscureText,
            autofocus: autofocus,
            decoration: InputDecoration(
              prefixIcon: const Icon(Icons.search_outlined),
              hintText: 'Search for someone to pay',
              hintStyle: GoogleFonts.inter(
                textStyle: const TextStyle(
                    color: TWColors.textMedium,
                    fontSize: 16,
                    fontWeight: FontWeight.w400),
              ),
              errorText: errorText,
            ),
          ),
        ],
      ),
    );
  }
}

class ExtendedTextFormField extends TextFormField {
  ExtendedTextFormField({super.key});
}
