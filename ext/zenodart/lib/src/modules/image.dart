import 'dart:io';
import 'package:image/image.dart' as img;
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class ImageModule {
  static void register(Executor executor) {
    executor.registerHandler('image.resize', _handleResize);
  }

  static Future<void> _handleResize(
      Node node, Scope scope, Executor executor) async {
    // Usage: image.resize: "input.png" { width: 100, height: 100, out: "output.png" }

    final inputRaw = executor.evaluateExpression(node.value, scope);
    final inputPath = inputRaw?.toString();

    if (inputPath == null || inputPath.isEmpty) return;

    final file = File(inputPath);
    if (!file.existsSync()) {
      print("Error: image file not found: $inputPath");
      return;
    }

    int? width;
    int? height;
    String? outPath;

    for (final child in node.children) {
      if (child.name == 'width') {
        final val = executor.evaluateExpression(child.value, scope);
        if (val is int)
          width = val;
        else if (val is String) width = int.tryParse(val);
      } else if (child.name == 'height') {
        final val = executor.evaluateExpression(child.value, scope);
        if (val is int)
          height = val;
        else if (val is String) height = int.tryParse(val);
      } else if (child.name == 'out') {
        final val = executor.evaluateExpression(child.value, scope);
        outPath = val?.toString();
      }
    }

    if (outPath == null) {
      print("Error: image.resize missing 'out' path");
      return;
    }

    try {
      final bytes = await file.readAsBytes();
      final image = img.decodeImage(bytes);

      if (image == null) {
        print("Error: failed to decode image $inputPath");
        return;
      }

      final resized = img.copyResize(image, width: width, height: height);
      final encoded = img.encodePng(resized);

      await File(outPath).writeAsBytes(encoded);
      print("Image resized: $inputPath -> $outPath");
    } catch (e) {
      print("image.resize error: $e");
    }
  }
}
