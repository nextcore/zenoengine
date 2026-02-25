# Filesystem & Uploads

Handling HTTP request uploads, storing files, and manipulating the filesystem in ZenoEngine revolves around the `http.upload` and built-in filesystem (`fs`) slots.

## Handling File Uploads

When receiving multipart/form-data POST requests containing files, you can use the `http.upload` slot. This slot simplifies the safe extraction and storage of uploaded media.

```zeno
http.post: '/api/profile/upload' {
    do: {
        // Retrieve the uploaded file using the input name "avatar"
        // And save it gracefully into the specified destination folder.
        http.upload: "avatar" {
            dest: "./storage/avatars"
            as: $uploadedFile
        }
        
        // $uploadedFile will contain metadata like original_name, size, and its saved path
        if: $uploadedFile == null {
            then: {
                http.bad_request: { error: "No file uploaded or invalid format!" }
                return
            }
        }
        
        // Save the file path to the database
        db.table: "users"
        db.where: "id" { equals: 1 }
        db.update: { avatar_path: $uploadedFile.path }

        http.ok: { message: "Avatar uploaded successfully!", file: $uploadedFile }
    }
}
```

The `http.upload` slot handles common potential errors securely. It restricts paths gracefully without risking directory traversal attacks while uploading.

## Basic Filesystem Operations

ZenoEngine provides simple OS-level integrations via Native filesystem operations like `fs.read`, `fs.write`, and checking if files exist (`fs.exists`). 

*Note: Since ZenoEngine targets web apps primarily, file operations are usually abstracted behind DB operations or cloud storage, and large file rendering is handled automatically by Blade/Views, or `http.static` instead.*

```zeno
// Reading the contents of a text file
fs.read: "storage/data/report.csv" {
    as: $csvData
}

// Checking if an important configuration exists
fs.exists: ".env" {
    as: $hasConfig
}

if: $hasConfig == false {
    then: {
        log: "WARNING: No .env configuration found!"
    }
}
```
