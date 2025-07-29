# Supabase Storage Integration

## Overview

This application now uses **Supabase Storage** for file management, replacing previous external APIs (file.io, MEGA, local storage). Supabase provides 1GB of free storage with automatic file management.

## Key Features

### ✅ Auto-Deletion After Admin Download
- When an admin downloads a file, it's **automatically deleted** from Supabase Storage
- This ensures optimal storage usage and prevents accumulation of files

### 📊 Admin Dashboard
- Real-time statistics (total files, storage used, active users)
- Advanced filtering by file type, user, and search terms
- Complete file management with download and delete capabilities

### 🔒 Security
- Files are stored in private buckets (not publicly accessible)
- Path validation prevents directory traversal attacks
- User-specific folder organization (`user_{userID}/`)

## Environment Variables

Add these to your `.env` file in the backend directory:

```env
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-anon-key-here
SUPABASE_STORAGE_BUCKET=uploads
```

## Getting Supabase Credentials

1. **Create a Supabase Account**:
   - Go to [supabase.com](https://supabase.com)
   - Create a new project

2. **Get Your Project URL**:
   - From your Supabase dashboard
   - Settings → API → Project URL

3. **Get Anon Key**:
   - Settings → API → Anon Key (public)
   - ✅ **Safe to use in client applications**

4. **Create Storage Bucket** (Optional):
   - The app will automatically create the `uploads` bucket
   - Or create it manually in Storage section

## File Structure

Files are organized as:
```
uploads/
├── user_123e4567-e89b-12d3-a456-426614174000/
│   ├── 123e4567_1672531200_document.pdf
│   └── 123e4567_1672531300_image.jpg
└── user_456f7890-f12c-34e5-b678-789123456789/
    └── 456f7890_1672531400_file.zip
```

## API Endpoints

### Admin Endpoints
- `GET /api/admin/file-upload/files` - List all files
- `GET /api/admin/file-upload/files/{userID}` - List user files  
- `GET /api/admin/file-upload/files/download/{filePath}` - Download & delete file
- `DELETE /api/admin/file-upload/files/{filePath}` - Delete file

### User Endpoints
- `POST /api/upload/file-authenticated` - Upload file (authenticated)
- `POST /api/upload/file` - Upload file with token

## File Upload Process

1. **Authenticated Upload**:
   ```javascript
   const formData = new FormData();
   formData.append('file', file);
   
   fetch('/api/upload/file-authenticated', {
     method: 'POST',
     headers: {
       'Authorization': `Bearer ${token}`,
     },
     body: formData,
   });
   ```

2. **Token-based Upload**:
   ```javascript
   const formData = new FormData();
   formData.append('file', file);
   formData.append('token', uploadToken);
   
   fetch('/api/upload/file', {
     method: 'POST',
     body: formData,
   });
   ```

## Storage Limits

- **Free Tier**: 1GB storage
- **File Size**: Configurable (currently no specific limit in code)
- **File Types**: PDF, DOC, DOCX, JPG, JPEG, PNG, GIF, TXT, ZIP, RAR

## Migration from Previous Systems

### Removed Services
- ❌ file.io integration (unreliable API)
- ❌ MEGA integration (payment required for API)  
- ❌ Local storage (server dependency)

### Benefits of Supabase Storage
- ✅ Reliable cloud storage
- ✅ 1GB free storage
- ✅ Built-in CDN
- ✅ Secure access controls
- ✅ Automatic file management
- ✅ No server disk usage

## Development

### Backend
```bash
cd backend
go build -o main .
./main
```

### Frontend
The admin file management page (`/admin/file-management`) provides:
- File statistics dashboard
- Advanced filtering options
- One-click download with auto-deletion
- File type icons and organization

## Troubleshooting

### Common Issues

1. **"Servicio de archivos no disponible"**
   - Check if `SUPABASE_URL` and `SUPABASE_KEY` are set
   - Verify Supabase project is active

2. **Bucket creation errors**
   - Ensure service role key has storage admin permissions
   - Check project billing status

3. **File upload failures**
   - Verify file type is allowed
   - Check file size limits
   - Ensure bucket exists and is accessible

### Debug Mode
Enable debug logging by adding:
```go
log.SetLevel(log.DebugLevel)
```

## Support

For issues related to:
- **Supabase**: Check [Supabase Documentation](https://supabase.com/docs)
- **Storage Limits**: Review your Supabase project usage
- **API Integration**: Verify environment variables and network connectivity

---

**Note**: This implementation automatically manages storage by deleting files after admin download, ensuring efficient use of the 1GB Supabase Storage limit. 