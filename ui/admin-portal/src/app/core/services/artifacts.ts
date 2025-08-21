import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ArtifactsService {
  private base = '/api/admin';

  constructor(private http: HttpClient) {}

  // Multipart upload: path + file
  uploadMultipart(repo: string, path: string, file: File): Observable<any> {
    const form = new FormData();
    form.append('path', path);
    form.append('file', file);
    return this.http.post(`${this.base}/repositories/${encodeURIComponent(repo)}/artifacts/upload`, form);
  }

  // Raw upload: send file/blob as body, path in query string
  uploadRaw(repo: string, path: string, data: Blob, contentType?: string): Observable<any> {
    const headers = new HttpHeaders(contentType ? { 'Content-Type': contentType } : {});
    return this.http.post(
      `${this.base}/repositories/${encodeURIComponent(repo)}/artifacts/upload?path=${encodeURIComponent(path)}`,
      data,
      { headers }
    );
  }

  // Move artifact: copy-then-delete operation handled on backend
  move(repo: string, fromPath: string, toPath: string): Observable<any> {
    const body = { from_path: fromPath, to_path: toPath } as any;
    return this.http.post(`${this.base}/repositories/${encodeURIComponent(repo)}/artifacts/move`, body);
  }

  // Copy artifact within the same repository
  copy(repo: string, fromPath: string, toPath: string): Observable<any> {
    const body = { from_path: fromPath, to_path: toPath } as any;
    return this.http.post(`${this.base}/repositories/${encodeURIComponent(repo)}/artifacts/copy`, body);
  }

  // Delete artifact via repository route
  delete(repo: string, path: string): Observable<any> {
    const encodedPath = path.split('/').map(encodeURIComponent).join('/');
    return this.http.delete(`/api/repositories/${encodeURIComponent(repo)}/${encodedPath}`);
  }
}
