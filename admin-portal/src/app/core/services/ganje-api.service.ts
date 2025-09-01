import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface RepositoryConfig {
  url?: string;
  description?: string;
  public?: boolean;
  enable_indexing?: boolean;
  [key: string]: any;
}

export interface Repository {
  id: number;
  name: string;
  type: string; // local, remote, virtual
  artifact_type: string;
  url?: string;
  config?: RepositoryConfig;
  created_at: string;
  updated_at: string;
  total_artifacts: number;
  total_size: number;
  pull_count: number;
  push_count: number;
}

export interface ArtifactInfo {
  id: number;
  repository_id: number;
  type: string;
  name: string;
  version: string;
  group?: string;
  path: string;
  size: number;
  checksum: string;
  yanked: boolean;
  created_at: string;
  updated_at: string;
  pull_count: number;
  push_count: number;
}

export interface CreateRepositoryRequest {
  name: string;
  type: string;
  artifact_type: string;
  url?: string;
  upstream?: string[];
  options?: { [key: string]: string };
  description?: string;
}

export interface RepositoryStats {
  total_artifacts: number;
  total_size: number;
  pull_count: number;
  push_count: number;
  artifact_types: { [key: string]: number };
}

export interface Webhook {
  id: number;
  repository_id: number;
  name: string;
  url: string;
  events: string;
  payload_template?: string;
  headers_json?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class GanjeApiService {
  private readonly baseUrl = '/api/v1';

  constructor(private http: HttpClient) {}

  // Repository Management
  getRepositories(): Observable<Repository[]> {
    return this.http.get<Repository[]>(`${this.baseUrl}/repositories`);
  }

  getRepository(name: string): Observable<Repository> {
    return this.http.get<Repository>(`${this.baseUrl}/repositories/${name}`);
  }

  createRepository(repository: CreateRepositoryRequest): Observable<Repository> {
    return this.http.post<Repository>(`${this.baseUrl}/repositories`, repository);
  }

  updateRepository(name: string, repository: Partial<Repository>): Observable<Repository> {
    return this.http.put<Repository>(`${this.baseUrl}/repositories/${name}`, repository);
  }

  deleteRepository(name: string, force: boolean = false): Observable<any> {
    const params = force ? new HttpParams().set('force', 'true') : new HttpParams();
    return this.http.delete(`${this.baseUrl}/repositories/${name}`, { params });
  }

  validateRepositoryConfig(config: CreateRepositoryRequest): Observable<any> {
    return this.http.post(`${this.baseUrl}/repositories/validate`, config);
  }

  getRepositoryTypes(): Observable<{ repository_types: string[], artifact_types: string[] }> {
    return this.http.get<{ repository_types: string[], artifact_types: string[] }>(`${this.baseUrl}/repositories/types`);
  }

  // Artifact Management
  getArtifacts(repositoryName: string): Observable<ArtifactInfo[]> {
    return this.http.get<ArtifactInfo[]>(`${this.baseUrl}/repositories/${repositoryName}/artifacts`);
  }

  getRepositoryStats(repositoryName: string): Observable<RepositoryStats> {
    return this.http.get<RepositoryStats>(`${this.baseUrl}/repositories/${repositoryName}/stats`);
  }

  downloadArtifact(repositoryName: string, path: string): Observable<Blob> {
    return this.http.get(`/${repositoryName}/${path}`, { responseType: 'blob' });
  }

  uploadArtifact(repositoryName: string, path: string, file: File): Observable<any> {
    return this.http.put(`/${repositoryName}/${path}`, file);
  }

  deleteArtifact(repositoryName: string, path: string): Observable<any> {
    return this.http.delete(`/${repositoryName}/${path}`);
  }

  // Cache Management
  invalidateCache(repositoryName: string, path?: string): Observable<any> {
    const params = path ? new HttpParams().set('path', path) : new HttpParams();
    return this.http.delete(`${this.baseUrl}/repositories/${repositoryName}/cache`, { params });
  }

  rebuildIndex(repositoryName: string): Observable<any> {
    return this.http.post(`${this.baseUrl}/repositories/${repositoryName}/reindex`, {});
  }

  // Webhook Management
  getWebhooks(repositoryName: string): Observable<Webhook[]> {
    return this.http.get<Webhook[]>(`${this.baseUrl}/repositories/${repositoryName}/webhooks`);
  }

  createWebhook(repositoryName: string, webhook: Partial<Webhook>): Observable<Webhook> {
    return this.http.post<Webhook>(`${this.baseUrl}/repositories/${repositoryName}/webhooks`, webhook);
  }

  updateWebhook(webhookId: number, updates: Partial<Webhook>): Observable<Webhook> {
    return this.http.put<Webhook>(`${this.baseUrl}/webhooks/${webhookId}`, updates);
  }

  deleteWebhook(webhookId: number): Observable<any> {
    return this.http.delete(`${this.baseUrl}/webhooks/${webhookId}`);
  }

  // Search
  searchRepositories(query: string): Observable<Repository[]> {
    const params = new HttpParams().set('q', query);
    return this.http.get<Repository[]>(`${this.baseUrl}/repositories/search`, { params });
  }

  searchArtifacts(query: string, repositoryName?: string): Observable<ArtifactInfo[]> {
    let params = new HttpParams().set('q', query);
    if (repositoryName) {
      params = params.set('repository', repositoryName);
    }
    return this.http.get<ArtifactInfo[]>(`${this.baseUrl}/artifacts/search`, { params });
  }
}
