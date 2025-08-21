import { TestBed } from '@angular/core/testing';

import { Artifacts } from './artifacts';

describe('Artifacts', () => {
  let service: Artifacts;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(Artifacts);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
