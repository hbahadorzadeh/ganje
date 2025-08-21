import { TestBed } from '@angular/core/testing';

import { Repositories } from './repositories';

describe('Repositories', () => {
  let service: Repositories;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(Repositories);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
