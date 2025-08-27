import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export enum PrivacyType {
  Public = 'public',
  Private = 'private',
}

export interface ArchiveDocument {
  doc_id: number;
  title: string;
  privacy: PrivacyType;
  created_at: string;
  created_by?: number | null;
  created_by_login?: string | null;
  created_by_full_name?: string | null;
  updated_at?: string | null;
  updated_by?: number | null;
  updated_by_login?: string | null;
  updated_by_full_name?: string | null;
  document_date?: string | null;
  author_id?: number | null;
  author_name?: string | null;
  type_id?: number | null;
  type_name?: string | null;
  tags?: string[];
  viewers?: number[];
  editors?: number[];
  can_requester_edit: boolean;
  geom?: string | null;
}

@Injectable()
export class DashboardService {
  private readonly http = inject(HttpClient);
  private readonly base = '/api/documents';


  getALLDocuments(): Observable<ArchiveDocument[]> {
    return this.http.get<ArchiveDocument[]>(this.base);
  }
}
