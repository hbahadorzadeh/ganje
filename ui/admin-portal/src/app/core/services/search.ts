import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface SearchResultItem {
  repository: string;
  path: string;
  size: number;
}

@Injectable({
  providedIn: 'root'
})
export class SearchService {
  private base = '/api/admin';

  constructor(private http: HttpClient) {}

  search(query: string, repository?: string): Observable<SearchResultItem[]> {
    let params = new HttpParams().set('q', query);
    if (repository) params = params.set('repository', repository);
    return this.http.get<SearchResultItem[]>(`${this.base}/search`, { params });
  }
}
