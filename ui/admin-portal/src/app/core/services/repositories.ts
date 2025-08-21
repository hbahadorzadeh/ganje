import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface RepositorySummary {
  name: string;
  artifactType: string;
  type?: string;
}

export interface ArtifactItem {
  path: string;
  size: number;
  checksum?: string;
  contentType?: string;
}

export interface RepoStats {
  totalArtifacts: number;
  totalSize: number;
}

@Injectable({
  providedIn: 'root'
})
export class RepositoriesService {
  private base = '/api/admin';

  constructor(private http: HttpClient) {}

  listRepositories(): Observable<RepositorySummary[]> {
    return this.http.get<RepositorySummary[]>(`${this.base}/repositories`);
  }

  getRepository(name: string): Observable<RepositorySummary> {
    return this.http.get<RepositorySummary>(`${this.base}/repositories/${encodeURIComponent(name)}`);
  }

  listArtifacts(name: string): Observable<ArtifactItem[]> {
    return this.http.get<ArtifactItem[]>(`${this.base}/repositories/${encodeURIComponent(name)}/artifacts`);
  }

  getStats(name: string): Observable<RepoStats> {
    return this.http.get<RepoStats>(`${this.base}/repositories/${encodeURIComponent(name)}/stats`);
  }
}
