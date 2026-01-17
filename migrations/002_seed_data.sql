-- ===========================================
-- Seed Data for Testing
-- ===========================================

-- Categories
INSERT INTO categories (name, slug) VALUES
    ('Hair Salon', 'hair-salon'),
    ('Barbershop', 'barbershop'),
    ('Nail Salon', 'nail-salon'),
    ('Spa', 'spa'),
    ('Beauty Salon', 'beauty-salon'),
    ('Skincare Clinic', 'skincare-clinic'),
    ('Massage Therapy', 'massage-therapy'),
    ('Tattoo & Piercing', 'tattoo-piercing');

-- Amenities
INSERT INTO amenities (name, icon) VALUES
    ('WiFi Gratis', 'wifi'),
    ('Estacionamiento', 'parking'),
    ('Acceso para sillas de ruedas', 'accessible'),
    ('Reserva Online', 'calendar'),
    ('Acepta Tarjetas', 'credit-card'),
    ('Sin turno previo', 'walk-in'),
    ('Salas Privadas', 'room'),
    ('Bebidas incluidas', 'coffee');

-- Sample SalonsS
INSERT INTO salons (name, slug, description, address, city, state, postal_code, country, latitude, longitude, phone, category_id, price_range, rating, review_count, is_verified) VALUES
    ('Estilo Mar', 'estilo-mar',
     'Peluquería premium especializada en coloración, mechas balayage y cortes de precisión. Nuestros estilistas expertos hacen realidad tu visión.',
     'Av. Colón 2145', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0023, -57.5575, '(0223) 495-0101', 1, 3, 4.8, 342, true),

    ('Barbería Don Pedro', 'barberia-don-pedro',
     'Barbería clásica con un toque moderno. Afeitadas con toalla caliente, arreglo de barba y cortes tradicionales en un ambiente relajado.',
     'San Martín 2380', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0055, -57.5428, '(0223) 495-0102', 2, 2, 4.6, 189, true),

    ('Spa Costa Atlántica', 'spa-costa-atlantica',
     'Spa de día completo que ofrece masajes, tratamientos faciales, corporales y servicios de bienestar holístico con vista al mar.',
     'Boulevard Marítimo 3200', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0156, -57.5337, '(0223) 495-0103', 4, 4, 4.9, 521, true),

    ('Nail Studio MDP', 'nail-studio-mdp',
     'Estudio creativo de uñas con gelificado, esculpidas, nail art y servicios de manicura y pedicura de lujo.',
     'Güemes 2890', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0012, -57.5512, '(0223) 495-0104', 3, 2, 4.4, 156, false),

    ('Tinta Urbana Tattoo', 'tinta-urbana-tattoo',
     'Artistas del tatuaje premiados, especializados en diseños personalizados, cover-ups y línea fina.',
     'Av. Independencia 1850', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0089, -57.5467, '(0223) 495-0105', 8, 3, 4.7, 278, true),

    ('Clínica Dermoestética Luz', 'clinica-dermoestetica-luz',
     'Tratamientos de skincare de grado médico: peelings químicos, microagujado, láser y soluciones anti-edad.',
     'Av. Luro 3456', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0034, -57.5523, '(0223) 495-0106', 6, 4, 4.5, 203, true),

    ('Corte Express', 'corte-express',
     'Cortes de pelo rápidos y accesibles para toda la familia. Sin turno previo, solo pasá!',
     'Av. Juan B. Justo 1200', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -37.9956, -57.5589, '(0223) 495-0107', 1, 1, 3.9, 89, false),

    ('Zen Masajes', 'zen-masajes',
     'Servicios de masajes terapéuticos: sueco, descontracturante, deportivo y aromaterapia.',
     'Rivadavia 2567', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0067, -57.5445, '(0223) 495-0108', 7, 2, 4.8, 312, true),

    ('La Porteña Barbershop', 'la-portena-barbershop',
     'Barbería estilo Buenos Aires con ambiente vintage. Fades, degradé y cortes frescos.',
     'Alem 2234', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0098, -57.5401, '(0223) 495-0109', 2, 2, 4.3, 145, true),

    ('Belleza Total', 'belleza-total',
     'Salón de belleza integral: peluquería, maquillaje, depilación y paquetes para novias.',
     'Catamarca 1876', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0045, -57.5534, '(0223) 495-0201', 5, 3, 4.6, 267, true),

    ('Nails & Co', 'nails-and-co',
     'Spa de uñas con productos orgánicos y ambiente relajante frente a la playa.',
     'Av. Peralta Ramos 450', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0234, -57.5298, '(0223) 495-0301', 3, 3, 4.5, 198, true),

    ('Peluquería Francesa', 'peluqueria-francesa',
     'Salón elegante en el centro. Especialistas en color creativo y estilos vanguardistas.',
     'Córdoba 1654', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0078, -57.5489, '(0223) 495-0401', 1, 3, 4.4, 234, true),

    ('Caballeros VIP', 'caballeros-vip',
     'Barbería ejecutiva para profesionales ocupados. Servicios premium y bebidas de cortesía.',
     'Av. Colón 3100', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0001, -57.5612, '(0223) 495-0402', 2, 3, 4.7, 456, true),

    ('Centro Estético Brillo', 'centro-estetico-brillo',
     'Centro de estética de lujo con tratamientos faciales europeos, dermaplaning y cuidados personalizados.',
     'Buenos Aires 2345', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0112, -57.5378, '(0223) 495-0403', 6, 4, 4.8, 189, true),

    ('Eco Hair Studio', 'eco-hair-studio',
     'Peluquería eco-friendly con productos sustentables. Especialistas en cabello natural y tratamientos orgánicos.',
     'Diagonal Pueyrredón 2678', 'Mar del Plata', 'Buenos Aires', '7600', 'Argentina', -38.0134, -57.5356, '(0223) 495-0501', 1, 2, 4.5, 176, true);

-- Sample Services
INSERT INTO services (salon_id, name, description, price_min, price_max, duration_minutes) VALUES
    -- Estilo Mar
    (1, 'Corte y Peinado', 'Corte de precisión con lavado y peinado', 8000, 15000, 60),
    (1, 'Balayage', 'Mechas pintadas a mano para un look natural', 25000, 45000, 180),
    (1, 'Color Completo', 'Coloración total con brillo final', 18000, 30000, 120),
    (1, 'Brushing', 'Lavado y secado profesional', 5000, 8000, 45),

    -- Barbería Don Pedro
    (2, 'Corte Clásico', 'Corte tradicional con tijera o máquina', 4000, 6000, 30),
    (2, 'Afeitada con Toalla Caliente', 'Afeitada de lujo con navaja y toalla caliente', 4500, 6000, 45),
    (2, 'Arreglo de Barba', 'Perfilado y recorte de barba', 2000, 3500, 20),
    (2, 'Servicio Completo', 'Corte, afeitada y arreglo de barba', 8000, 11000, 75),

    -- Spa Costa Atlántica
    (3, 'Masaje Sueco', 'Masaje relajante de cuerpo completo', 15000, 22000, 60),
    (3, 'Masaje Descontracturante', 'Masaje terapéutico para aliviar tensiones', 18000, 25000, 60),
    (3, 'Facial Signature', 'Tratamiento facial personalizado con productos premium', 20000, 28000, 75),
    (3, 'Paquete Parejas', 'Experiencia de masaje lado a lado', 35000, 50000, 90);

-- Salon Amenities
INSERT INTO salon_amenities (salon_id, amenity_id) VALUES
    (1, 1), (1, 4), (1, 5), (1, 7),  -- Estilo Mar
    (2, 1), (2, 5), (2, 6), (2, 8),  -- Barbería Don Pedro
    (3, 1), (3, 2), (3, 3), (3, 4), (3, 5), (3, 7), (3, 8),  -- Spa Costa Atlántica
    (4, 1), (4, 4), (4, 5), (4, 6),  -- Nail Studio MDP
    (5, 4), (5, 5),                   -- Tinta Urbana
    (6, 1), (6, 2), (6, 3), (6, 4), (6, 5), (6, 7);  -- Clínica Dermoestética

-- Operating Hours (sample for first few salons)
INSERT INTO operating_hours (salon_id, day_of_week, open_time, close_time, is_closed) VALUES
    -- Estilo Mar (closed Sunday, Monday)
    (1, 0, NULL, NULL, true),
    (1, 1, NULL, NULL, true),
    (1, 2, '10:00', '19:00', false),
    (1, 3, '10:00', '19:00', false),
    (1, 4, '10:00', '20:00', false),
    (1, 5, '09:00', '20:00', false),
    (1, 6, '09:00', '18:00', false),

    -- Barbería Don Pedro (open every day)
    (2, 0, '10:00', '17:00', false),
    (2, 1, '09:00', '19:00', false),
    (2, 2, '09:00', '19:00', false),
    (2, 3, '09:00', '19:00', false),
    (2, 4, '09:00', '20:00', false),
    (2, 5, '09:00', '20:00', false),
    (2, 6, '08:00', '18:00', false);
