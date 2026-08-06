package com.personal.repositories;

import com.personal.entities.BuildEntity;
import org.springframework.data.jpa.repository.JpaRepository;

public interface BuildRepository extends JpaRepository<BuildEntity, String> {}
