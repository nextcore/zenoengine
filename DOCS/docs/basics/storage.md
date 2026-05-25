# File Storage

ZenoEngine provides a unified local file storage abstraction through the `storage.*` slot group, making it easy to store, delete, and check files on the local filesystem.

## Configuration

ZenoEngine supports multiple storage drivers (local filesystem and S3/S3-compatible cloud storage like MinIO, Cloudflare R2, or AWS S3) via the `.env` file.

### 1. Local Filesystem Configuration (Default)

By default, the storage system saves files to the public assets directory (`./public/storage`) so that saved files can be served directly over HTTP. You can customize the root storage directory:

```env
STORAGE_DISK=local
STORAGE_DIR=public/storage
```

### 2. S3 / S3-Compatible Cloud Storage Configuration

To store files directly in AWS S3 or any S3-compatible API (MinIO, Cloudflare R2, DigitalOcean Spaces, Wasabi, etc.):

```env
STORAGE_DISK=s3
S3_ENDPOINT=s3.amazonaws.com     # Or play.min.io, CLOUDFLARE_ACCOUNT_ID.r2.cloudflarestorage.com, etc.
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=your_bucket_name
S3_REGION=us-east-1              # Optional region
S3_SSL=true                      # Set to false for local dev MinIO HTTP endpoints
```

> [!NOTE]
> If the S3 configuration is incomplete (e.g. during local development), ZenoEngine will automatically log a warning and fall back to the `local` filesystem driver so that your application remains fully functional without crashing.

---

## Storing Files (`storage.put`)

The `storage.put` slot writes file contents or copies existing files into the configured storage directory. It automatically creates any missing parent directories.

### 1. Saving Plain Text Content
Great for generating text files, JSON configs, or CSV logs dynamically:

```zl
storage.put {
  content: 'ID,Name,Email\n1,John,john@example.com'
  path: 'reports/users_today.csv'
  as: $saved_path
}
```

### 2. Saving Raw Binary/Byte Content
Useful when receiving binary data from APIs or file generators:

```zl
storage.put {
  content: $binary_image_data
  path: 'avatars/user_101.png'
  as: $avatar_url
}
```

### 3. Copying Existing Uploaded Files
When handling file uploads via `http.upload` (which saves files to a temporary directory), you can copy the file to its final destination in storage:

```zl
# 1. Capture the temporary uploaded file from the request
http.upload {
  field: 'avatar'
  dest: 'public/uploads/tmp'
  as: $temp_filename
}

# 2. Copy/move the temporary file to its organized storage destination
if $temp_filename != '' {
  storage.put {
    content: 'public/uploads/tmp/' + $temp_filename
    path: 'users/avatars/' + $temp_filename
    is_file_path: true
    as: $final_path
  }
}
```

### Parameter Reference for `storage.put`

* **`content`** (string or bytes, **Required**): The content to write, or the source file path if `is_file_path` is true.
* **`path`** (string, **Required**): Target relative path within the storage directory.
* **`is_file_path`** (bool, Optional): Set to `true` to treat `content` as a source file path to copy from. Defaults to `false`.
* **`as`** (string, Optional): Variable to store the saved relative path (Defaults to `$storage_path`).

---

## Deleting Files (`storage.delete`)

To delete a file from storage, use the `storage.delete` slot:

```zl
storage.delete: 'users/avatars/old_avatar.png' {
  as: $is_deleted
}
```

If the file does not exist, it returns `false` instead of throwing an error.

---

## Checking File Existence (`storage.exists`)

To verify if a file exists in the storage directory, use the `storage.exists` slot:

```zl
storage.exists: 'reports/users_today.csv' {
  as: $report_exists
}
```

It returns `true` if the file exists and is not a directory; otherwise, it returns `false`.

---

## Path Traversal Security

To prevent malicious users from reading or deleting files outside the allowed directory (e.g. attempting to input `../../etc/passwd`), the storage system cleans and checks paths using Go's `filepath.Clean`. 

If any path attempts to reference folders above the storage root, the slot will abort immediately and raise a **`storage: path traversal attempt detected`** security error.
