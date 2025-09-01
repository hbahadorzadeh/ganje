import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

export interface LoadingState {
  [key: string]: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class LoadingService {
  private loadingSubject = new BehaviorSubject<LoadingState>({});
  public loading$ = this.loadingSubject.asObservable();

  private currentState: LoadingState = {};

  setLoading(key: string, loading: boolean): void {
    this.currentState = {
      ...this.currentState,
      [key]: loading
    };
    this.loadingSubject.next(this.currentState);
  }

  isLoading(key: string): Observable<boolean> {
    return new Observable(observer => {
      this.loading$.subscribe(state => {
        observer.next(!!state[key]);
      });
    });
  }

  isAnyLoading(): Observable<boolean> {
    return new Observable(observer => {
      this.loading$.subscribe(state => {
        observer.next(Object.values(state).some(loading => loading));
      });
    });
  }

  clearLoading(key: string): void {
    const newState = { ...this.currentState };
    delete newState[key];
    this.currentState = newState;
    this.loadingSubject.next(this.currentState);
  }

  clearAllLoading(): void {
    this.currentState = {};
    this.loadingSubject.next(this.currentState);
  }
}
