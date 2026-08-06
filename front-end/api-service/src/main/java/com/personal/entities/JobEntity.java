package com.personal.entities;

import jakarta.persistence.*;

@Entity
public class JobEntity {
    @Id @GeneratedValue(strategy = GenerationType.UUID)
    private String id;
    private String name;

    @ManyToOne
    @JoinColumn(name = "repository_id")
    private RepositoryEntity repository;

    public String getId() { return id; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public RepositoryEntity getRepository() { return repository; }
    public void setRepository(RepositoryEntity repository) { this.repository = repository; }
}
